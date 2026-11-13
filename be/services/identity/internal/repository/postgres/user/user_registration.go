package userpostgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

func (r *Repository) CreateUserWithSession(
	ctx context.Context,
	input userDomain.RegistrationParams,
	session model.Session,
) (model.User, error) {
	// Parse identifiers before opening the transaction.
	userID, err := uuid.Parse(input.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("parse user ID: %w", err)
	}

	sessionID, err := uuid.Parse(session.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("parse session ID: %w", err)
	}

	var createdUser model.User

	// Persist the user and session atomically.
	err = postgresTransaction.WithinTransaction(
		ctx,
		r.pool,
		func(tx pgx.Tx) error {
			queries := r.queries.WithTx(tx)

			// Create the user record and map duplicate email errors.
			created, err := queries.CreateUser(ctx, db.CreateUserParams{
				ID:           userID,
				Email:        input.Email,
				PasswordHash: input.PasswordHash,
				FullName:     input.FullName,
				Phone:        input.Phone,
			})
			if err != nil {
				if identityException.IsUserEmailConflict(err) {
					return identityException.ErrAuthEmailTaken
				}

				return fmt.Errorf("create user: %w", err)
			}

			// Create the refresh session for the new user.
			sessionParams := postgresHelper.MapSessionParams(
				sessionID,
				userID,
				session,
			)
			if err := queries.CreateAuthSession(ctx, sessionParams); err != nil {
				return fmt.Errorf("create registration session: %w", err)
			}

			createdUser = postgresHelper.MapUser(created)
			return nil
		},
	)
	if err != nil {
		return model.User{}, err
	}

	return createdUser, nil
}
