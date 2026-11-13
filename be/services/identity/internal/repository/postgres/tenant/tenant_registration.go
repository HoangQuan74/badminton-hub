package tenantpostgres

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

func (r *Repository) ListRegistrations(ctx context.Context, status string) ([]model.TenantRegistration, error) {
	rows, err := r.queries.ListTenantRegistrations(ctx, status)
	if err != nil {
		return nil, err
	}

	items := make([]model.TenantRegistration, 0, len(rows))
	for _, row := range rows {
		items = append(items, registrationModel(db.GetTenantRegistrationRow(row)))
	}

	return items, nil
}

func registrationModel(row db.GetTenantRegistrationRow) model.TenantRegistration {
	return model.TenantRegistration{
		Organization: organizationModel(
			row.ID,
			row.Name,
			row.Slug,
			row.Status,
			row.CreatedAt.Time,
			row.UpdatedAt.Time,
		),
		ContactName:    row.ContactName,
		ContactEmail:   row.ContactEmail,
		ContactPhone:   row.ContactPhone,
		Note:           row.Note,
		ReviewDecision: row.ReviewDecision,
		ReviewNote:     row.ReviewNote,
		ReviewerID:     row.ReviewerID,
		ReviewedAt:     row.ReviewedAt.Time,
	}
}
