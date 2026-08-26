package main

import (
	"context"
	"time"
)

type apiDatabase interface {
	// upserts the update based on idempotency key, returns update id.
	upsertUpdate(ctx context.Context, update *upsertUpdate) (string, error)
	getUpdateById(ctx context.Context, id string) (*update, error)
	getUpdateLatest(ctx context.Context, base, quote string) (*update, error)
}

type workerDatabase interface {
	// returns updates that have a "pending" status.
	// TODO: add count property
	getPendingUpdates(ctx context.Context, count int) ([]*pendingUpdate, error)
	saveUpdateResults(ctx context.Context, successes []*successfulUpdate, failures []string) error
}

type updateStatus string

const (
	updateStatusPending   updateStatus = "pending"
	updateStatusCompleted updateStatus = "completed"
	updateStatusFailed    updateStatus = "failed"
)

func (s updateStatus) String() string {
	return string(s)
}

type update struct {
	Id        string
	Base      string
	Quote     string
	Status    updateStatus
	Price     *float64
	UpdatedAt *time.Time
}

type upsertUpdate struct {
	IdempotencyKey string
	Base           string
	Quote          string
}

type pendingUpdate struct {
	Id    string
	Base  string
	Quote string
}

type successfulUpdate struct {
	Id        string
	Price     float64
	UpdatedAt time.Time
}
