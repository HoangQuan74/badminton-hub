package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TimestampPointer converts an optional time filter to a PostgreSQL value.
func TimestampPointer(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{
		Time:  *value,
		Valid: true,
	}
}

// WithinTransaction executes work in one PostgreSQL transaction using the
// default isolation settings. The callback owns SQLC query binding.
func WithinTransaction(ctx context.Context, pool *pgxpool.Pool, work func(pgx.Tx) error) error {
	return WithinTransactionWithOptions(ctx, pool, pgx.TxOptions{}, work)
}

// WithinTransactionWithOptions executes work in one PostgreSQL transaction and
// allows a repository to choose isolation or access mode when required.
func WithinTransactionWithOptions(
	ctx context.Context,
	pool *pgxpool.Pool,
	options pgx.TxOptions,
	work func(pgx.Tx) error,
) error {
	tx, err := pool.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := work(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
