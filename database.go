package main

import (
	"context"
	"time"
)

type apiDatabase interface {
	createUpdate(ctx context.Context, update *newUpdate) error
	getUpdateById(ctx context.Context, id string) (*update, error)
	getUpdateLatest(ctx context.Context, base, quote string) (*update, error)
}

type workerDatabase interface {
	getPendingUpdates(ctx context.Context) ([]*pendingUpdate, error)
	completeUpdates(ctx context.Context, updates []*completeUpdate) error
}

// TODO: idempotency
// TODO: error handling...
type update struct {
	Id        string
	Base      string
	Quote     string
	Status    string
	Price     *float64
	UpdatedAt *time.Time
}

type newUpdate struct {
	Id    string
	Base  string
	Quote string
}

type pendingUpdate struct {
	Id    string
	Base  string
	Quote string
}

type completeUpdate struct {
	Id        string
	Price     float64
	UpdatedAt time.Time
}
