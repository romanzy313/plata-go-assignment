package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
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

	e := newHTTPServer(logger, api, db)
	sc := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", cfg.Port),
		GracefulTimeout: 10 * time.Second,
		HideBanner:      true,
	}
	if err := sc.Start(ctx, e); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}
