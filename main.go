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
	var _ = cfg

	deps := &handler{
		exchangeRateClient: nil,
		apiDatabase:        nil,
	}
	e := newHTTPServer(deps)

	e.Start(fmt.Sprintf(":%d", cfg.Port))
}
