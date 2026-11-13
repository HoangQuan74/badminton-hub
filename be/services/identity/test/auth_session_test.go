package identitytest

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type refreshRepository struct {
	auth.AuthRepository
	transaction *refreshTransaction
	commitError error
}

func (r *refreshRepository) WithinSessionTransaction(_ context.Context, execute func(auth.AuthSessionTransaction) error) error {
	if err := execute(r.transaction); err != nil {
		return err
	}

	return r.commitError
}

type refreshTransaction struct {
	principal    model.User
	lookupError  error
	persistError error
	revocations  int
	creations    int
	session      model.Session
}

func (t *refreshTransaction) LockSessionOwner(context.Context, []byte, time.Time) (string, error) {
	return t.principal.ID, nil
}

func (t *refreshTransaction) GetPrincipal(context.Context, string) (model.User, error) {
	return t.principal, t.lookupError
}

func (t *refreshTransaction) RevokeSession(context.Context, []byte, time.Time) error {
	t.revocations++
	return nil
}

func (t *refreshTransaction) CreateSession(_ context.Context, session model.Session) error {
	t.creations++
	t.session = session
	return t.persistError
}

func TestRefreshRejectsAccountBeforeConsumingSession(t *testing.T) {
	lookupError := errors.New("principal lookup failed")
	cases := []struct {
		name        string
		status      model.AccountStatus
		role        string
		policy      auth.AccessPolicy
		lookupError error
		wantError   error
	}{
		{
			name:      "inactive owner",
			status:    constant.AccountStatusSuspended,
			role:      constant.RoleTenantOwner,
			policy:    tenant.CanAccessPortal,
			wantError: identityException.ErrAuthUserInactive,
		},
		{
			name:      "wrong portal",
			status:    constant.AccountStatusActive,
			role:      constant.RoleCustomer,
			policy:    tenant.CanAccessPortal,
			wantError: identityException.ErrAuthWrongPortal,
		},
		{
			name:      "missing policy fails closed",
			status:    constant.AccountStatusActive,
			role:      constant.RoleTenantOwner,
			wantError: identityException.ErrAuthWrongPortal,
		},
		{
			name:        "unavailable principal",
			lookupError: lookupError,
			wantError:   lookupError,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			transaction := &refreshTransaction{
				principal: model.User{
					ID:       uuid.NewString(),
					Status:   test.status,
					RoleName: test.role,
				},
				lookupError: test.lookupError,
			}
			tokens := auth.NewTokenManager("test-secret", time.Minute, time.Hour)
			service := auth.NewService(&refreshRepository{transaction: transaction}, tokens, test.policy)
			pair, err := service.Refresh(context.Background(), "existing-refresh-token")
			if !errors.Is(err, test.wantError) {
				t.Fatalf("refresh error: %v, want %v", err, test.wantError)
			}
			if transaction.revocations != 0 || transaction.creations != 0 || pair != (model.TokenPair{}) {
				t.Fatal("denied refresh must not consume the session or expose new tokens")
			}
		})
	}
}

func TestRefreshPreparesMatchingResponseAndStoredSession(t *testing.T) {
	transaction := &refreshTransaction{
		principal: model.User{
			ID:       uuid.NewString(),
			Status:   constant.AccountStatusActive,
			RoleName: constant.RoleTenantOwner,
		},
	}
	tokens := auth.NewTokenManager("test-secret", time.Minute, time.Hour)
	repository := &refreshRepository{transaction: transaction}
	service := auth.NewService(repository, tokens, tenant.CanAccessPortal)
	pair, err := service.Refresh(context.Background(), "existing-refresh-token")
	if err != nil {
		t.Fatal(err)
	}
	if transaction.revocations != 1 || transaction.creations != 1 {
		t.Fatal("refresh must replace exactly one session")
	}
	principalID, err := tokens.ParseAccessToken(pair.AccessToken)
	if err != nil || principalID != transaction.principal.ID || transaction.session.UserID != principalID {
		t.Fatalf("response and storage must identify the same principal: %v", err)
	}
	if !bytes.Equal(transaction.session.TokenHash, auth.HashRefreshToken(pair.RefreshToken)) {
		t.Fatal("persisted hash does not match the returned refresh token")
	}
	if !transaction.session.ExpiresAt.Equal(pair.RefreshExpiresAt) {
		t.Fatal("persisted and returned refresh expiry must match")
	}
	if _, err := uuid.Parse(transaction.session.ID); err != nil {
		t.Fatalf("invalid replacement session identifier: %v", err)
	}

	// Persistence failure must not expose a usable replacement token pair.
	transaction.persistError = errors.New("session write failed")
	pair, err = service.Refresh(context.Background(), "existing-refresh-token")
	if !errors.Is(err, transaction.persistError) || pair != (model.TokenPair{}) {
		t.Fatal("failed replacement must return its error without new tokens")
	}

	transaction.persistError = nil
	repository.commitError = errors.New("session commit failed")
	pair, err = service.Refresh(context.Background(), "existing-refresh-token")
	if !errors.Is(err, repository.commitError) || pair != (model.TokenPair{}) {
		t.Fatal("failed commit must return its error without new tokens")
	}
}
