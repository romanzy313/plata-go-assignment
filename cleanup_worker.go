package main

import (
	"context"
	"log/slog"
	"time"
)

type cleanupWorker struct {
	logger        *slog.Logger
	db            workerDatabase
	pollInterval  time.Duration
	staleDuration time.Duration
}

// Cleanup Worker marks stale updates as failed in the database.
// Currently, it works without pagination pagination.
func newCleanupWorker(
	logger *slog.Logger,
	db workerDatabase,
	pollInterval time.Duration,
	staleDuration time.Duration,
) *cleanupWorker {
	return &cleanupWorker{
		logger:        logger.With("source", "cleanupWorker"),
		db:            db,
		pollInterval:  pollInterval,
		staleDuration: staleDuration,
	}
}

func (w *cleanupWorker) Run(ctx context.Context) {
	w.processWithLogging(ctx)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processWithLogging(ctx)
		}
	}
}

func (w *cleanupWorker) processWithLogging(ctx context.Context) {
	cleanupCount, err := w.Process(ctx)
	if err != nil {
		w.logger.Error("cleanup worker error", "error", err)
		return
	}
	if cleanupCount == 0 {
		w.logger.Debug("no updates where cleaned up")
		return
	}

	w.logger.Debug("cleaned up updates", "count", cleanupCount)
}

func (w *cleanupWorker) Process(ctx context.Context) (int64, error) {
	return w.db.failStaleUpdates(ctx, w.staleDuration)
}
