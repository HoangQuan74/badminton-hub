package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithinTransaction owns begin, commit, and rollback; callers bind their SQLC queries.
func WithinTransaction(ctx context.Context, pool *pgxpool.Pool, work func(pgx.Tx) error) error {
	return WithinTransactionWithOptions(
		ctx,
		pool,
		pgx.TxOptions{},
		work,
	)
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
