package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	ctx := context.TODO()

	cfg, err := newConfigFromEnv()
	if err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	db, err := newDatabasePostgres(ctx, cfg.DatabaseUrl)
	if err != nil {
		slog.Error("Database error", "error", err)
		os.Exit(1)
	}
	defer db.close(ctx)

	rateClient := newExchangeratesapiClient(cfg.ExchangeratesapiKey)

	handler := &handler{
		exchangeRateClient: rateClient,
		apiDatabase:        db,
	}

	e := newHTTPServer(handler)

	e.Start(fmt.Sprintf(":%d", cfg.Port))
}
