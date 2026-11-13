package exception

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TenantPersistenceError(err error) error {
	// Preserve cancellation and domain failures before classifying dependencies.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		if postgresError.Code == "23505" && postgresError.ConstraintName == "identity_organizations_slug_unique_idx" {
			return ErrTenantSlugConflict
		}
		if strings.HasPrefix(postgresError.Code, "08") || postgresError.Code == "57P01" || postgresError.Code == "57P02" || postgresError.Code == "57P03" {
			return fmt.Errorf("%w: %w", ErrTenantUnavailable, err)
		}
	}
	var connectError *pgconn.ConnectError
	var networkError net.Error
	if errors.As(err, &connectError) || errors.As(err, &networkError) {
		return fmt.Errorf("%w: %w", ErrTenantUnavailable, err)
	}
	return err
}

func TenantWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "identity_organizations_slug_unique_idx" {
		return ErrTenantSlugConflict
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "identity_venues_slug_unique_idx" {
		return ErrTenantVenueSlugConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTenantNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
