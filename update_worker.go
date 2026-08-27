package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type updateWorker struct {
	logger        *slog.Logger
	api           exchangeRateClient
	db            workerDatabase
	batchSize     int
	pollInterval  time.Duration
	staleDuration time.Duration

	snapshot          *rateSnapshot
	snapshotFetchedAt time.Time
}

func newUpdateWorker(
	logger *slog.Logger,
	api exchangeRateClient,
	db workerDatabase,
	batchSize int,
	pollInterval time.Duration,
	staleDuration time.Duration,
) *updateWorker {
	return &updateWorker{
		logger:        logger.With("source", "updateWorker"),
		api:           api,
		db:            db,
		batchSize:     batchSize,
		pollInterval:  pollInterval,
		staleDuration: staleDuration,
	}
}

func (w *updateWorker) Run(ctx context.Context) {
	w.processUntilEmpty(ctx)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processUntilEmpty(ctx)
		}
	}
}

func (w *updateWorker) processUntilEmpty(ctx context.Context) {
	for {
		// stop when context is cancelled
		if ctx.Err() != nil {
			return
		}

		more, err := w.Process(ctx)
		if err != nil {
			w.logger.Error("worker error", "error", err)
			return
		}
		if !more {
			return
		}
	}
}

func (w *updateWorker) currentSnapshot(
	ctx context.Context,
) (*rateSnapshot, error) {
	if w.snapshot != nil &&
		time.Now().Before(w.snapshotFetchedAt.Add(w.staleDuration)) {
		return w.snapshot, nil
	}

	snapshot, err := w.api.latestSnapshot(ctx, supportedCurrencies)
	if err != nil {
		return nil, err
	}

	w.snapshot = snapshot
	w.snapshotFetchedAt = time.Now().UTC()

	return snapshot, nil
}

// Process returns true if more updates are pending
func (w *updateWorker) Process(ctx context.Context) (bool, error) {
	pendingUpdates, err := w.db.getPendingUpdates(ctx, w.batchSize, w.staleDuration)
	if err != nil {
		return false, err
	}
	if len(pendingUpdates) == 0 {
		w.logger.Debug("no pending updates")
		return false, nil
	}

	successfulUpdates := make([]*successfulUpdate, 0, len(pendingUpdates))
	failedUpdateIds := make([]string, 0)

	w.logger.Debug("processing updates", "count", len(pendingUpdates))

	snapshot, err := w.currentSnapshot(ctx)
	if err != nil {
		for _, pending := range pendingUpdates {
			failedUpdateIds = append(failedUpdateIds, pending.Id)
		}

		if saveErr := w.db.saveUpdateResults(
			ctx,
			successfulUpdates,
			failedUpdateIds,
		); saveErr != nil {
			return false, saveErr
		}

		return false, fmt.Errorf("failed to get exchange rates: %w", err)
	}

	var lastError error
	for _, pending := range pendingUpdates {
		rate, err := calculateExchangeRate(
			snapshot,
			pending.Base,
			pending.Quote,
		)
		if err != nil {
			failedUpdateIds = append(failedUpdateIds, pending.Id)
			lastError = err
			continue
		}

		successfulUpdates = append(successfulUpdates, &successfulUpdate{
			Id:        pending.Id,
			Price:     rate,
			UpdatedAt: snapshot.UpdatedAt,
		})
	}

	if err := w.db.saveUpdateResults(
		ctx,
		successfulUpdates,
		failedUpdateIds,
	); err != nil {
		return false, err
	}

	w.logger.Debug(
		"update results",
		"successes", len(successfulUpdates),
		"failures", len(failedUpdateIds),
		"calculationErr", lastError,
	)

	return true, nil
}
