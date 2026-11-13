package tenant

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type Service struct{ repository TenantRepository }

func NewService(repository TenantRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListOrganizations(ctx context.Context, actorID string) ([]model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, identityException.ErrTenantForbidden
	}

	return s.repository.ListOrganizations(ctx, actorID)
}

func (s *Service) GetOrganization(ctx context.Context, actorID, id string) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(id) == "" {
		return model.Organization{}, identityException.ErrTenantInvalidInput
	}

	return s.repository.GetOrganization(
		ctx,
		actorID,
		id,
	)
}

func (s *Service) UpdateOrganization(ctx context.Context, actorID, id string, input UpdateOrganizationInput) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(id) == "" || (input.Name == nil && input.Slug == nil) {
		return model.Organization{}, identityException.ErrTenantInvalidInput
	}

	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !utils.IsRequiredText(value, 120) {
			return model.Organization{}, identityException.ErrTenantInvalidInput
		}

		input.Name = &value
	}

	if input.Slug != nil {
		value := utils.NormalizeLowercase(*input.Slug)
		if !utils.IsSlug(value, 100) {
			return model.Organization{}, identityException.ErrTenantInvalidInput
		}

		input.Slug = &value
	}

	// Authorize owner-only organization changes before persistence.
	if err := s.authorizeOrganizationOwner(
		ctx,
		actorID,
		id,
	); err != nil {
		return model.Organization{}, err
	}

	return s.repository.UpdateOrganization(
		ctx,
		actorID,
		id,
		input,
	)
}

func (s *Service) ArchiveOrganization(ctx context.Context, actorID, id string) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(id) == "" {
		return model.Organization{}, identityException.ErrTenantInvalidInput
	}

	// Authorize owner-only archival before the conditional write.
	if err := s.authorizeOrganizationOwner(
		ctx,
		actorID,
		id,
	); err != nil {
		return model.Organization{}, err
	}

	return s.repository.ArchiveOrganization(
		ctx,
		actorID,
		id,
	)
}

func (s *Service) CreateOrganization(ctx context.Context, actorID string, input CreateOrganizationInput) (CreateOrganizationOutput, error) {
	// Validate the server-derived actor and the complete input.
	if _, err := uuid.Parse(actorID); err != nil {
		return CreateOrganizationOutput{}, identityException.ErrTenantUnauthorized
	}

	if err := validation.Struct(input); err != nil {
		return CreateOrganizationOutput{}, identityException.TenantInputError(err)
	}

	// Canonicalize business values before hashing or persistence.
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = utils.NormalizeLowercase(input.Slug)
	input.IdempotencyKey = strings.ToLower(input.IdempotencyKey)
	input.ContactName = strings.TrimSpace(input.ContactName)
	input.ContactEmail = utils.NormalizeLowercase(input.ContactEmail)
	input.ContactPhone = strings.TrimSpace(input.ContactPhone)
	input.Note = strings.TrimSpace(input.Note)
	fields := []identityException.TenantFieldViolation{}
	if input.Name == "" {
		fields = append(fields, identityException.TenantFieldViolation{
			Field: "name",
			Code:  "REQUIRED",
		})
	} else if strings.ContainsRune(input.Name, '\x00') {
		fields = append(fields, identityException.TenantFieldViolation{
			Field: "name",
			Code:  "NAME_INVALID",
		})
	}

	if !utils.IsSlug(input.Slug, 100) {
		fields = append(fields, identityException.TenantFieldViolation{
			Field: "slug",
			Code:  "SLUG_INVALID",
		})
	}

	if len(fields) > 0 {
		return CreateOrganizationOutput{}, &identityException.TenantValidationError{Fields: fields}
	}

	// Hash only canonical business fields; transport metadata is excluded.
	payload, err := json.Marshal([6]string{
		input.Name,
		input.Slug,
		input.ContactName,
		input.ContactEmail,
		input.ContactPhone,
		input.Note,
	})
	if err != nil {
		return CreateOrganizationOutput{}, err
	}

	hash := sha256.Sum256(payload)
	params := CreateOrganizationParams{
		ActorID:     actorID,
		Name:        input.Name,
		Slug:        input.Slug,
		Key:         input.IdempotencyKey,
		RequestHash: hash[:],
	}
	var output CreateOrganizationOutput
	err = s.repository.WithinCreationTransaction(ctx, func(tx TenantCreationTransaction) error {
		// Recheck eligibility against the locked, current account.
		actor, err := tx.LockCreator(ctx, actorID)
		if err != nil {
			return err
		}

		if err := CheckActor(actor); err != nil {
			return err
		}

		// Return the committed result only when the same business request is replayed.
		replay, found, err := tx.FindCreateReplay(
			ctx,
			actorID,
			input.IdempotencyKey,
		)
		if err != nil {
			return err
		}

		if found {
			if !bytes.Equal(replay.RequestHash, params.RequestHash) {
				return identityException.ErrTenantKeyConflict
			}

			output = replay.Output
			output.Replayed = true
			return nil
		}

		// Apply owner enrollment and create the pending organization together.
		if err := tx.DeleteExpiredCreateKey(
			ctx,
			actorID,
			input.IdempotencyKey,
		); err != nil {
			return err
		}

		if err := tx.SetOwnerRole(ctx, actorID); err != nil {
			return err
		}

		organization, err := tx.CreateOrganization(ctx, params)
		if err != nil {
			return err
		}

		actor.RoleName = constant.RoleTenantOwner
		output = CreateOrganizationOutput{
			User:         actor,
			Organization: organization,
			Membership: model.OrganizationStaff{
				OrganizationID: organization.ID,
				UserID:         actor.ID,
				Role:           constant.RoleTenantOwner,
				Status:         constant.MembershipStatusActive,
			},
		}

		// Capture the requested contact, falling back to the existing account details.
		contact := RegistrationContactInput{
			ContactName:  input.ContactName,
			ContactEmail: input.ContactEmail,
			ContactPhone: input.ContactPhone,
			Note:         input.Note,
		}
		if contact.ContactName == "" {
			contact.ContactName = actor.FullName
		}

		if contact.ContactEmail == "" {
			contact.ContactEmail = actor.Email
		}

		if contact.ContactPhone == "" {
			contact.ContactPhone = actor.Phone
		}

		if err := tx.SaveRegistration(
			ctx,
			organization.ID,
			contact,
		); err != nil {
			return err
		}

		// Save the replay result before committing any enrollment effect.
		expiresAt, err := tx.SaveCreateReplay(
			ctx,
			params,
			output,
		)
		if err != nil {
			return err
		}

		output.ExpiresAt = expiresAt
		return nil
	})
	if err != nil {
		return CreateOrganizationOutput{}, err
	}

	return output, nil
}

// CheckActor is reapplied to the current database row inside the transaction.
func CheckActor(user model.User) error {
	if user.Status != constant.AccountStatusActive || user.IsAdmin {
		return identityException.ErrTenantForbidden
	}

	switch user.RoleName {
	case "", constant.RoleCustomer, constant.RoleTenantOwner:
		return nil
	default:
		return identityException.ErrTenantRoleConflict
	}
}

func (s *Service) ListRegistrations(ctx context.Context, status string) ([]model.TenantRegistration, error) {
	switch status {
	case "", constant.OrganizationStatusPendingVerification, constant.OrganizationStatusActive, constant.OrganizationStatusRejected, constant.OrganizationStatusSuspended, constant.OrganizationStatusArchived:
		return s.repository.ListRegistrations(ctx, status)
	default:
		return nil, identityException.ErrTenantInvalidInput
	}
}

func (s *Service) ReviewRegistration(ctx context.Context, actorID string, input ReviewRegistrationInput) (model.TenantRegistration, error) {
	// Validate review identity and the requested decision.
	if _, err := uuid.Parse(input.OrganizationID); err != nil {
		return model.TenantRegistration{}, identityException.ErrTenantInvalidInput
	}

	if _, err := uuid.Parse(actorID); err != nil {
		return model.TenantRegistration{}, identityException.ErrTenantUnauthorized
	}

	if (input.Decision != constant.OrganizationStatusActive && input.Decision != constant.OrganizationStatusRejected) || utf8.RuneCountInString(input.Note) > 1000 || strings.ContainsRune(input.Note, '\x00') {
		return model.TenantRegistration{}, identityException.ErrTenantInvalidInput
	}

	input.Note = strings.TrimSpace(input.Note)

	// Decide the state transition while the registration is locked.
	var result model.TenantRegistration
	err := s.repository.WithinReviewTransaction(ctx, func(tx TenantReviewTransaction) error {
		current, err := tx.LockRegistration(ctx, input.OrganizationID)
		if err != nil {
			return err
		}

		changed, err := PlanRegistrationReview(current.Organization.Status, input.Decision)
		if err != nil {
			return err
		}

		if !changed {
			result = current
			return nil
		}

		// Persist the accepted transition and audit record in the same transaction.
		if err := tx.SaveRegistrationReview(
			ctx,
			actorID,
			input,
		); err != nil {
			return err
		}

		result, err = tx.GetRegistration(ctx, input.OrganizationID)
		return err
	})
	return result, err
}

func CanAccessPortal(user model.User) bool {
	if user.IsAdmin {
		return false
	}

	return user.RoleName == constant.RoleTenantOwner || user.RoleName == constant.RoleTenantStaff || user.RoleName == constant.RoleVenueStaff
}

// PlanRegistrationReview owns the transition decision; the service calls it under a row lock.
func PlanRegistrationReview(current, decision string) (bool, error) {
	if decision != constant.OrganizationStatusActive && decision != constant.OrganizationStatusRejected {
		return false, identityException.ErrTenantInvalidInput
	}

	if current == decision {
		return false, nil
	}

	if current != constant.OrganizationStatusPendingVerification {
		return false, identityException.ErrTenantStateConflict
	}

	return true, nil
}

func (s *Service) authorizeOrganizationOwner(ctx context.Context, actorID, organizationID string) error {
	role, err := s.repository.GetOrganizationRole(
		ctx,
		actorID,
		organizationID,
	)
	if err != nil {
		return err
	}

	if role != constant.RoleTenantOwner {
		return identityException.ErrTenantForbidden
	}

	return nil
}
