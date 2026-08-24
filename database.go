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
	// returns updates that have a "pending" status
	getPendingUpdates(ctx context.Context) ([]*pendingUpdate, error)
	completeUpdates(ctx context.Context, updates []*completeUpdate) error
	failUpdates(ctx context.Context, updateIds []string) error
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
	Id             string
	IdempotencyKey string
	Base           string
	Quote          string
	Status         updateStatus
	Price          *float64
	UpdatedAt      *time.Time
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
