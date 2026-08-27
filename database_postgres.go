package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// https://knrdl.github.io/posts/postgres-enum/
const schema = `
DO
$$
    BEGIN
		CREATE TYPE update_status AS ENUM ('pending', 'processing', 'completed', 'failed');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

CREATE TABLE IF NOT EXISTS updates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  idempotency_key UUID NOT NULL UNIQUE,
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

type databasePostgres struct {
	pool *pgxpool.Pool
}

func newDatabasePostgres(ctx context.Context, databaseURL string) (*databasePostgres, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}
	config.MaxConns = 2

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	d := &databasePostgres{pool: conn}
	if err := d.migrate(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return d, nil
}

func (d *databasePostgres) close() {
	d.pool.Close()
}

func (d *databasePostgres) migrate(ctx context.Context) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, schema)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	return err
}

func (d *databasePostgres) upsertUpdate(ctx context.Context, update *upsertUpdate) (string, error) {
	var id string
	err := d.pool.QueryRow(ctx, `
INSERT INTO updates (
  idempotency_key,
  base_currency,
  quote_currency,
  status
)
VALUES (
  $1,
  $2,
  $3,
  'pending'
)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING id::text;`, update.IdempotencyKey, update.Base, update.Quote).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (d *databasePostgres) getUpdateById(ctx context.Context, id string) (*update, error) {
	var u update
	err := d.pool.QueryRow(ctx, `
SELECT
  id::text,
  base_currency,
  quote_currency,
  status,
  price,
  updated_at
FROM updates
WHERE id = $1
LIMIT 1;
`, id).Scan(&u.Id, &u.Base, &u.Quote, &u.Status, &u.Price, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (d *databasePostgres) getUpdateLatest(ctx context.Context, base, quote string) (*update, error) {
	u := update{
		Base:   base,
		Quote:  quote,
		Status: updateStatusCompleted,
	}
	err := d.pool.QueryRow(ctx, `
SELECT 
  id::text,
  price,
  updated_at
FROM updates
WHERE
  base_currency = $1 AND
  quote_currency = $2 AND
  status = 'completed'
ORDER BY updated_at DESC
LIMIT 1;
`, base, quote).Scan(&u.Id, &u.Price, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// https://habr.com/ru/articles/984102/
func (d *databasePostgres) getPendingUpdates(ctx context.Context, count int) ([]*pendingUpdate, error) {
	rows, err := d.pool.Query(ctx, `
WITH update_selection AS (
  SELECT id
  FROM updates
  WHERE status = 'pending'
  ORDER BY created_at ASC
  LIMIT $1
  FOR NO KEY UPDATE SKIP LOCKED
)
UPDATE updates AS u
SET status = 'processing'
FROM update_selection AS selected
WHERE
  u.id = selected.id AND
  u.status = 'pending'
RETURNING u.id::text, u.base_currency, u.quote_currency;
`, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pendingUpdates := make([]*pendingUpdate, 0)
	for rows.Next() {
		var u pendingUpdate
		if err := rows.Scan(&u.Id, &u.Base, &u.Quote); err != nil {
			return nil, err
		}
		pendingUpdates = append(pendingUpdates, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pendingUpdates, nil
}

func (d *databasePostgres) saveUpdateResults(
	ctx context.Context,
	successes []*successfulUpdate,
	failures []string) error {
	if len(successes) == 0 && len(failures) == 0 {
		return nil
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// save success
	for _, u := range successes {
		res, err := tx.Exec(ctx, `
UPDATE updates
SET
  status = 'completed',
  price = $2,
  updated_at = $3
WHERE
  id = $1 AND
  status = 'processing';`, u.Id, u.Price, u.UpdatedAt)
		if err != nil {
			return err
		}
		if res.RowsAffected() != 1 {
			return fmt.Errorf("could not complete pending update %s", u.Id)
		}
	}

	// save failures
	if len(failures) > 0 {
		_, err = tx.Exec(ctx, `
UPDATE updates
SET
  status = 'failed'
WHERE
  id = ANY($1) AND
  status = 'pending';`, failures)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
