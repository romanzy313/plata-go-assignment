package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type worker struct {
	logger       *slog.Logger
	api          exchangeRateClient
	db           workerDatabase
	batchSize    int
	pollInterval time.Duration
	// TODO: items older then this are changed to failed
	// pendingTimeout time.Duration
}

func newWorker(
	logger *slog.Logger,
	api exchangeRateClient,
	db workerDatabase,
	pollInterval time.Duration,
) *worker {
	return &worker{
		logger:       logger,
		api:          api,
		db:           db,
		batchSize:    3, // for demo
		pollInterval: pollInterval,
	}
}

func (w *worker) Run(ctx context.Context) {
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

func (w *worker) processUntilEmpty(ctx context.Context) {
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

// Process returns true if more updates are pending
func (w *worker) Process(ctx context.Context) (bool, error) {
	pendingUpdates, err := w.db.getPendingUpdates(ctx, w.batchSize)
	if err != nil {
		return false, err
	}
	if len(pendingUpdates) == 0 {
		w.logger.Debug("no pending updates")
		return false, nil
	}

	successfulUpdates := make([]*successfulUpdate, 0, len(pendingUpdates))
	failedUpdateIds := make([]string, 0)
	var lastError error = nil

	w.logger.Debug("processing updates", "count", len(pendingUpdates))
	snapshot, err := w.api.latestSnapshot(ctx, supportedCurrencies)
	if err == nil {
		for _, pending := range pendingUpdates {
			rate, err := calculateExchangeRate(snapshot, pending.Base, pending.Quote)
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
	} else {
		lastError = err
		for _, pending := range pendingUpdates {
			failedUpdateIds = append(failedUpdateIds, pending.Id)
		}
	}

	err = w.db.saveUpdateResults(ctx, successfulUpdates, failedUpdateIds)
	w.logger.Debug("update results",
		"successes", len(successfulUpdates),
		"errors", len(failedUpdateIds),
		"dbError", err,
		"lastError", lastError)
	if err != nil {
		return false, err
	}

	if lastError != nil {
		return false, fmt.Errorf("some calculations failed: %d failures; last err: %w",
			len(failedUpdateIds),
			lastError)
	}

	return len(pendingUpdates) > 0, nil
}
