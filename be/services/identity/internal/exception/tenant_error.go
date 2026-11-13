package exception

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var (
	ErrTenantUnauthorized      = errors.New("tenant.unauthorized")
	ErrTenantRoleConflict      = errors.New("tenant.role_conflict")
	ErrTenantKeyConflict       = errors.New("tenant.idempotency_conflict")
	ErrTenantUnavailable       = errors.New("tenant.unavailable")
	ErrTenantInvalidInput      = errors.New("invalid tenant input")
	ErrTenantForbidden         = errors.New("tenant access forbidden")
	ErrTenantNotFound          = errors.New("tenant resource not found")
	ErrTenantVenueSlugConflict = errors.New("venue slug already exists")
	ErrTenantSlugConflict      = errors.New("organization slug already exists")
	ErrTenantStateConflict     = errors.New("organization state does not allow this operation")
)

type TenantFieldViolation struct {
	Field string
	Code  string
}

type TenantValidationError struct{ Fields []TenantFieldViolation }

func (e *TenantValidationError) Error() string { return ErrTenantInvalidInput.Error() }
func (e *TenantValidationError) Unwrap() error { return ErrTenantInvalidInput }

func TenantInputError(err error) error {
	var violations validator.ValidationErrors
	if !errors.As(err, &violations) {
		return ErrTenantInvalidInput
	}
	fields := make([]TenantFieldViolation, 0, len(violations))
	for _, violation := range violations {
		field := map[string]string{"Name": "name", "Slug": "slug", "IdempotencyKey": "idempotency_key"}[violation.StructField()]
		code := map[string]string{"required": "REQUIRED", "max": "TOO_LONG", "uuid4_rfc4122": "UUID_INVALID"}[violation.Tag()]
		fields = append(fields, TenantFieldViolation{Field: field, Code: code})
	}
	return &TenantValidationError{Fields: fields}
}

var ErrTenantOwnerRoleMissing = errors.New("owner system role is missing")
