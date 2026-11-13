package auth

import (
	"context"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type Service struct {
	policy AccessPolicy
	repo   AuthRepository
	tokens *TokenManager
	now    func() time.Time
}

func NewService(repo AuthRepository, tokens *TokenManager, policy AccessPolicy) *Service {
	return &Service{
		policy: policy,
		repo:   repo,
		tokens: tokens,
		now:    time.Now,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (model.User, model.TokenPair, error) {
	// Validate the complete login input.
	if err := validation.Struct(input); err != nil {
		return model.User{}, model.TokenPair{}, identityException.ErrAuthInvalidCredentials
	}

	// Canonicalize the email before lookup.
	input.Email = utils.NormalizeLowercase(input.Email)

	// Load the user and verify the password.
	found, err := s.repo.FindCredentialsByEmail(ctx, input.Email)
	if err != nil || !security.ComparePassword(found.PasswordHash, input.Password) {
		return model.User{}, model.TokenPair{}, identityException.ErrAuthInvalidCredentials
	}

	// Enforce the current account status and portal policy before issuing a session.
	if err := s.validatePrincipal(found); err != nil {
		return model.User{}, model.TokenPair{}, err
	}

	// Issue a new authenticated session.
	pair, err := s.issueSession(ctx, found.ID)
	return found, pair, err
}

func (s *Service) Refresh(ctx context.Context, oldToken string) (model.TokenPair, error) {
	// Reject empty credentials before preparing a replacement session.
	if oldToken == "" {
		return model.TokenPair{}, identityException.ErrAuthInvalidToken
	}

	tokenHash := HashRefreshToken(oldToken)

	// Validate current account admission before consuming the locked session.
	var output model.TokenPair
	err := s.repo.WithinSessionTransaction(ctx, func(tx AuthSessionTransaction) error {
		owner, err := tx.LockSessionOwner(
			ctx,
			tokenHash,
			s.now().UTC(),
		)
		if err != nil {
			return err
		}

		principal, err := tx.GetPrincipal(ctx, owner)
		if err != nil {
			return err
		}

		if err := s.validatePrincipal(principal); err != nil {
			return err
		}

		// Prepare both tokens before making the old session unusable.
		pair, replacement, err := s.tokens.NewSession(owner)
		if err != nil {
			return err
		}

		// Commit revocation and replacement as one authentication operation.
		if err := tx.RevokeSession(
			ctx,
			tokenHash,
			s.now().UTC(),
		); err != nil {
			return err
		}

		if err := tx.CreateSession(ctx, replacement); err != nil {
			return err
		}

		output = pair
		return nil
	})
	if err != nil {
		return model.TokenPair{}, err
	}

	return output, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	// Validate the refresh token before revoking the session.
	if refreshToken == "" {
		return identityException.ErrAuthInvalidToken
	}

	// Revoke the session identified by the hashed refresh token.
	return s.repo.RevokeSession(
		ctx,
		HashRefreshToken(refreshToken),
		s.now().UTC(),
	)
}

func (s *Service) GetPrincipal(ctx context.Context, userID string) (model.User, error) {
	return s.repo.GetPrincipal(ctx, userID)
}

func (s *Service) ParseAccessToken(raw string) (string, error) {
	return s.tokens.ParseAccessToken(raw)
}

func (s *Service) issueSession(ctx context.Context, userID string) (model.TokenPair, error) {
	// Prepare the response and hashed session through the shared token issuer.
	pair, session, err := s.tokens.NewSession(userID)
	if err != nil {
		return model.TokenPair{}, err
	}

	// Persist only after the complete token pair is ready.
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return model.TokenPair{}, err
	}

	return pair, nil
}

// AccessPolicy is supplied by the domain which owns admission to the caller's portal.
type AccessPolicy func(model.User) bool

func (s *Service) CurrentPrincipal(ctx context.Context, accessToken string) (model.User, error) {
	// Verify the token and load the current account state.
	id, err := s.tokens.ParseAccessToken(accessToken)
	if err != nil {
		return model.User{}, err
	}

	principal, err := s.repo.GetPrincipal(ctx, id)
	if err != nil {
		return model.User{}, err
	}

	if err := s.validatePrincipal(principal); err != nil {
		return model.User{}, err
	}

	return principal, nil
}

// validatePrincipal applies the same admission checks to login, refresh, and current user.
func (s *Service) validatePrincipal(principal model.User) error {
	if principal.Status != constant.AccountStatusActive {
		return identityException.ErrAuthUserInactive
	}

	if s.policy == nil || !s.policy(principal) {
		return identityException.ErrAuthWrongPortal
	}

	return nil
}
