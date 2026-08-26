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
		pollInterval: pollInterval,
	}
}

func (w *worker) Run(ctx context.Context) {
	if err := w.Process(ctx); err != nil {
		w.logger.Error("worker error", "error", err)
	}
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.Process(ctx); err != nil {
				w.logger.Error("worker error", "error", err)
			}
		}
	}
}

func (w *worker) Process(ctx context.Context) error {
	pendingUpdates, err := w.db.getPendingUpdates(ctx)
	if err != nil {
		return err
	}
	if len(pendingUpdates) == 0 {
		w.logger.Debug("no pending updates")
		return nil
	}

	w.logger.Debug("processing updates", "count", len(pendingUpdates))
	snapshot, err := w.api.latestSnapshot(ctx, supportedCurrencies)
	if err != nil {
		// FIXME: mark as failed
		return err
	}

	successfulUpdates := make([]*successfulUpdate, 0, len(pendingUpdates))
	failedUpdateIds := make([]string, 0)
	var lastError error = nil

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

	err = w.db.saveUpdateResults(ctx, successfulUpdates, failedUpdateIds)
	w.logger.Debug("update results",
		"successes", len(successfulUpdates),
		"errors", len(failedUpdateIds),
		"dbError", err,
		"lastError", lastError)
	if err != nil {
		return err
	}

	if lastError != nil {
		return fmt.Errorf("calculations failed: %d failures; last err: %w",
			len(failedUpdateIds),
			lastError)
	}

	return nil
}
