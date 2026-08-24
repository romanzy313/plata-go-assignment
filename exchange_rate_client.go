package main

import (
	"context"
	"time"
)

type exchangeRateClient interface {
	latestSnapshot(ctx context.Context, symbols []string) (rateSnapshot, error)
}

type rateSnapshot struct {
	Rates     map[string]float64
	UpdatedAt time.Time
}
