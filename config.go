package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type config struct {
	DatabaseUrl         string
	ExchangeratesapiKey string
	Port                int
	WorkerPollInterval  time.Duration
}

func newConfigFromEnv() (*config, error) {
	cfg := config{
		DatabaseUrl:         "",
		ExchangeratesapiKey: "",
		Port:                3000,
		WorkerPollInterval:  time.Minute,
	}

	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, fmt.Errorf("missing DATABASE_URL")
	}
	cfg.DatabaseUrl = databaseUrl

	apiKey, ok := os.LookupEnv("EXCHANGERATESAPI_KEY")
	if !ok {
		return nil, fmt.Errorf("missing EXCHANGERATESAPI_KEY")
	}
	cfg.ExchangeratesapiKey = apiKey

	if portStr, ok := os.LookupEnv("PORT"); ok {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("incorrect PORT: %w", err)
		}
		cfg.Port = port
	}

	if workerPollStr, ok := os.LookupEnv("WORKER_POLL_INTERVAL"); ok {
		workerPollInterval, err := time.ParseDuration(workerPollStr)
		if err != nil {
			return nil, fmt.Errorf("incorrect WORKER_POLL_INTERVAL: %w", err)
		}
		cfg.WorkerPollInterval = workerPollInterval
	}

	return &cfg, nil
}
