package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		slog.Error("Error loading .env file", "error", err)
		os.Exit(1)
	}

	deps := &handler{
		exchangeRateClient: nil,
		apiDatabase:        nil,
	}
	e := newHTTPServer(deps)

	e.Start(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
