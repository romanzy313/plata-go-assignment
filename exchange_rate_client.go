package main

import (
	"context"
	"time"
)

type exchangeRateClient interface {
	latestSnapshot(ctx context.Context, base string, symbols []string) (*rateSnapshot, error)
}

type rateSnapshot struct {
	Base      string
	Rates     map[string]float64
	UpdatedAt time.Time
}
