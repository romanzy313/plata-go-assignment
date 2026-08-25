package main

import (
	"context"
)

type exchangeRateClient interface {
	latestSnapshot(ctx context.Context, base string, symbols []string) (*rateSnapshot, error)
}
