package tenant

import (
	"context"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type TenantRepository interface {
	GetOrganizationRole(ctx context.Context, actorID, organizationID string) (string, error)
	ListOrganizations(ctx context.Context, actorID string) ([]model.Organization, error)
	GetOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error)
	UpdateOrganization(ctx context.Context, actorID, organizationID string, input UpdateOrganizationInput) (model.Organization, error)
	ArchiveOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error)
	ListRegistrations(ctx context.Context, status string) ([]model.TenantRegistration, error)
	WithinCreationTransaction(ctx context.Context, execute func(TenantCreationTransaction) error) error
	WithinReviewTransaction(ctx context.Context, execute func(TenantReviewTransaction) error) error
}

type TenantCreationTransaction interface {
	LockCreator(ctx context.Context, actorID string) (model.User, error)
	FindCreateReplay(ctx context.Context, actorID, key string) (CreationReplay, bool, error)
	DeleteExpiredCreateKey(ctx context.Context, actorID, key string) error
	SetOwnerRole(ctx context.Context, actorID string) error
	CreateOrganization(ctx context.Context, params CreateOrganizationParams) (model.Organization, error)
	SaveRegistration(ctx context.Context, organizationID string, input RegistrationContactInput) error
	SaveCreateReplay(ctx context.Context, params CreateOrganizationParams, output CreateOrganizationOutput) (time.Time, error)
}

type TenantReviewTransaction interface {
	LockRegistration(ctx context.Context, organizationID string) (model.TenantRegistration, error)
	SaveRegistrationReview(ctx context.Context, actorID string, input ReviewRegistrationInput) error
	GetRegistration(ctx context.Context, organizationID string) (model.TenantRegistration, error)
}

type CreateOrganizationParams struct {
	ActorID     string
	Name        string
	Slug        string
	Key         string
	RequestHash []byte
}

type CreationReplay struct {
	Output      CreateOrganizationOutput
	RequestHash []byte
}
