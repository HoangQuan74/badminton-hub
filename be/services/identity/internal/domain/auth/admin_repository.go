package auth

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type AdminRepository interface {
	FindAdminByEmail(context.Context, string) (model.AdminCredentials, error)
	CreateSession(context.Context, Session) error
	RotateAdminSession(context.Context, []byte, Session) (string, error)
	RevokeAdminSession(context.Context, []byte) error
	PasswordHash(context.Context, string) (string, error)
	UpdatePassword(context.Context, string, string) error
}
