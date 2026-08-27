package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const batchSize = 3

func main() {
	if err := run(); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := newConfigFromEnv()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	db, err := newDatabasePostgres(ctx, cfg.DatabaseUrl)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer db.close()

	api := newExchangeratesapiClient(cfg.ExchangeratesapiKey)

	worker := newUpdateWorker(logger, api, db, batchSize, cfg.WorkerPollInterval, cfg.StaleUpdateDuration)
	go worker.Run(ctx)

	cleanupWorker := newCleanupWorker(logger, db, cfg.WorkerPollInterval, cfg.StaleUpdateDuration)
	go cleanupWorker.Run(ctx)

	server := newServer(logger, db)
	server.Run(ctx, cfg.Port)

	return nil
}
