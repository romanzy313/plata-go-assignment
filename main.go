package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	cfg, err := newConfigFromEnv()
	if err != nil {
		slog.Error("Error reading config", "error", err)
		os.Exit(1)
	}
	rateClient := newExchangeratesapiClient(cfg.ExchangeratesapiKey)

	handler := &handler{
		exchangeRateClient: rateClient,
		apiDatabase:        nil,
	}
	e := newHTTPServer(handler)

	e.Start(fmt.Sprintf(":%d", cfg.Port))
}
