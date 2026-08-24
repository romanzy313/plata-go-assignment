package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type databasePostgres struct {
	conn *pgx.Conn
}

func newDatabasePostgres(ctx context.Context, databaseURL string) (*databasePostgres, error) {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}
	d := &databasePostgres{conn: conn}
	if err := d.migrate(ctx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return d, nil
}

func (d *databasePostgres) close(ctx context.Context) {
	d.conn.Close(ctx)
}

func (d *databasePostgres) migrate(ctx context.Context) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// https://knrdl.github.io/posts/postgres-enum/
	const schema = `
DO
$$
    BEGIN
		CREATE TYPE update_status AS ENUM ('pending', 'completed', 'failed');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

CREATE TABLE IF NOT EXISTS updates (
  id UUID PRIMARY KEY,
  idempotency_key TEXT NOT NULL UNIQUE,
  base_currency CHAR(3) NOT NULL,
  quote_currency CHAR(3) NOT NULL,
  status update_status NOT NULL,
  price DOUBLE PRECISION,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS updates_pending_idx
  ON updates (created_at)
  WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS updates_latest_idx
  ON updates (base_currency, quote_currency, updated_at DESC)
  WHERE status = 'completed';`
	_, err = tx.Exec(ctx, schema)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	return err
}

func (d *databasePostgres) createUpdate(ctx context.Context, update *newUpdate) error {
	return fmt.Errorf("Not implemented")
}

func (d *databasePostgres) getUpdateById(ctx context.Context, id string) (*update, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (d *databasePostgres) getUpdateLatest(ctx context.Context, base, quote string) (*update, error) {
	return nil, fmt.Errorf("Not implemented")
}
