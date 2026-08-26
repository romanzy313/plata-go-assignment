package main

import (
	"context"
)

type exchangeRateClient interface {
	latestSnapshot(ctx context.Context, symbols []string) (*rateSnapshot, error)
}
