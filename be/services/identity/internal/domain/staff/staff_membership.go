package staff

import (
	"context"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
)

type MembershipService struct{ repository StaffMembershipRepository }

func NewMembershipService(repository StaffMembershipRepository) *MembershipService {
	return &MembershipService{repository: repository}
}

func (s *MembershipService) ListOrganizationStaff(ctx context.Context, actorID, organizationID string) ([]model.OrganizationStaffDetails, error) {
	if _, err := uuid.Parse(organizationID); err != nil {
		return nil, identityException.ErrTenantInvalidInput
	}

	return s.repository.ListOrganizationStaff(
		ctx,
		actorID,
		organizationID,
	)
}

func (s *MembershipService) AddOrganizationStaff(ctx context.Context, actorID string, input AddOrganizationStaffInput) (model.OrganizationStaff, error) {
	// Validate the organization scope and requested role.
	if _, err := uuid.Parse(input.OrganizationID); err != nil {
		return model.OrganizationStaff{}, identityException.ErrTenantInvalidInput
	}

	if input.Role != constant.RoleTenantStaff && input.Role != constant.RoleVenueStaff {
		return model.OrganizationStaff{}, identityException.ErrTenantInvalidInput
	}

	// Normalize the exact account address before persistence.
	input.Email = utils.NormalizeLowercase(input.Email)
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email {
		return model.OrganizationStaff{}, identityException.ErrTenantInvalidInput
	}

	// Resolve the current organization/account state inside the write transaction.
	var output model.OrganizationStaff
	err = s.repository.WithinMembershipTransaction(ctx, func(tx StaffMembershipTransaction) error {
		status, err := tx.GetManagedOrganizationStatus(
			ctx,
			actorID,
			input.OrganizationID,
		)
		if err != nil {
			return err
		}

		if status != constant.OrganizationStatusActive {
			return identityException.ErrTenantForbidden
		}

		candidate, err := tx.LockCandidate(ctx, input.Email)
		if err != nil {
			return err
		}

		// Apply staff admission and promotion rules before changing either record.
		if err := ValidateMembershipCandidate(candidate, input.Role); err != nil {
			return err
		}

		if err := tx.AssignRole(
			ctx,
			candidate.ID,
			input.Role,
		); err != nil {
			return err
		}

		output, err = tx.SetOrganizationStaff(
			ctx,
			actorID,
			SetOrganizationStaffInput{
				OrganizationID: input.OrganizationID,
				UserID:         candidate.ID,
				Role:           input.Role,
				Status:         constant.MembershipStatusActive,
			},
		)
		return err
	})
	return output, err
}

func (s *MembershipService) RemoveOrganizationStaff(ctx context.Context, actorID string, input RemoveOrganizationStaffInput) error {
	// Validate both resource identifiers before requesting revocation.
	if _, err := uuid.Parse(input.OrganizationID); err != nil {
		return identityException.ErrTenantInvalidInput
	}

	if _, err := uuid.Parse(input.UserID); err != nil {
		return identityException.ErrTenantInvalidInput
	}

	// Revoke both organization access and venue assignments as one business operation.
	_, err := s.deactivateOrganizationStaff(
		ctx,
		actorID,
		input,
	)
	return err
}

func (s *MembershipService) SetOrganizationStaff(ctx context.Context, actorID string, input SetOrganizationStaffInput) (model.OrganizationStaff, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.OrganizationStaff{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(input.OrganizationID) == "" || strings.TrimSpace(input.UserID) == "" ||
		(input.Role != constant.RoleTenantStaff && input.Role != constant.RoleVenueStaff) ||
		(input.Status != constant.MembershipStatusActive && input.Status != constant.MembershipStatusInactive) {
		return model.OrganizationStaff{}, identityException.ErrTenantInvalidInput
	}

	// Reuse the same revocation flow as explicit member removal.
	if input.Status == constant.MembershipStatusInactive {
		return s.deactivateOrganizationStaff(
			ctx,
			actorID,
			RemoveOrganizationStaffInput{
				OrganizationID: input.OrganizationID,
				UserID:         input.UserID,
			},
		)
	}

	// Persist the validated active membership.
	return s.repository.SetOrganizationStaff(
		ctx,
		actorID,
		input,
	)
}

func (s *MembershipService) SetVenueStaff(ctx context.Context, actorID string, input SetVenueStaffInput) (model.VenueStaff, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.VenueStaff{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(input.VenueID) == "" || strings.TrimSpace(input.UserID) == "" ||
		(input.Status != constant.MembershipStatusActive && input.Status != constant.MembershipStatusInactive) {
		return model.VenueStaff{}, identityException.ErrTenantInvalidInput
	}

	return s.repository.SetVenueStaff(
		ctx,
		actorID,
		input,
	)
}

// ValidateMembershipCandidate owns eligibility for the requested staff role.
func ValidateMembershipCandidate(candidate model.User, role string) error {
	if role != constant.RoleTenantStaff && role != constant.RoleVenueStaff {
		return identityException.ErrTenantInvalidInput
	}

	if candidate.Status != constant.AccountStatusActive || candidate.IsAdmin {
		return identityException.ErrTenantRoleConflict
	}

	switch candidate.RoleName {
	case "", constant.RoleCustomer, role:
		return nil
	default:
		return identityException.ErrTenantRoleConflict
	}
}

// deactivateOrganizationStaff keeps membership and venue revocation atomic for both callers.
func (s *MembershipService) deactivateOrganizationStaff(ctx context.Context, actorID string, input RemoveOrganizationStaffInput) (model.OrganizationStaff, error) {
	var output model.OrganizationStaff
	err := s.repository.WithinMembershipTransaction(ctx, func(tx StaffMembershipTransaction) error {
		// Revoke organization membership first, retaining its final state for the response.
		var err error
		output, err = tx.DeactivateOrganizationStaff(
			ctx,
			actorID,
			input,
		)
		if err != nil {
			return err
		}

		// Remove all venue access before the transaction can commit.
		return tx.DeactivateMemberVenues(
			ctx,
			input.OrganizationID,
			input.UserID,
		)
	})
	if err != nil {
		return model.OrganizationStaff{}, err
	}

	return output, nil
}
