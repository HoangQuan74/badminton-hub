package helper

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

// MapUser accepts the common account projection used by create, update, and list.
// SQLC rows with the same shape are converted explicitly at the call site.
func MapUser(row db.CreateUserRow) model.User {
	return model.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FullName:     row.FullName,
		Phone:        row.Phone,
		Status:       model.AccountStatus(row.Status),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

// MapUserByID converts the wider lookup row to the same public account projection.
func MapUserByID(row db.FindUserByIDRow) model.User {
	return MapUser(db.CreateUserRow{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FullName:     row.FullName,
		Phone:        row.Phone,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	})
}

// MapPrincipal adds the access fields available in credential and principal queries.
func MapPrincipal(row db.FindUserByIDRow) model.User {
	user := MapUserByID(row)
	user.RoleName = row.RoleName
	user.IsAdmin = row.IsAdmin
	return user
}
