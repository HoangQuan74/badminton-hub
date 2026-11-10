package catalogpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/postgres/generated/sqlc"
)

type Repository struct{ queries *db.Queries }

func New(pool *pgxpool.Pool) *Repository { return &Repository{queries: db.New(pool)} }

func (r *Repository) ListCourts(ctx context.Context, venueID, organizationID string, activeOnly bool) ([]model.Court, error) {
	venue, err := parseID(venueID)
	if err != nil {
		return nil, err
	}
	organization, err := parseID(organizationID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCourtsByVenue(ctx, db.ListCourtsByVenueParams{VenueID: venue, OrganizationID: organization, ActiveOnly: activeOnly})
	if err != nil {
		return nil, fmt.Errorf("list courts: %w", err)
	}
	courts := make([]model.Court, 0, len(rows))
	for _, row := range rows {
		courts = append(courts, courtModel(row.ID, row.VenueID, row.OrganizationID, row.Name, row.Status, row.SlotIntervalMinutes, row.CreatedAt.Time, row.UpdatedAt.Time))
	}
	return courts, nil
}
func (r *Repository) GetCourt(ctx context.Context, courtID string) (model.Court, error) {
	id, err := parseID(courtID)
	if err != nil {
		return model.Court{}, err
	}
	row, err := r.queries.GetCourt(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Court{}, catalog.ErrNotFound
	}
	if err != nil {
		return model.Court{}, fmt.Errorf("get court: %w", err)
	}
	return courtModel(row.ID, row.VenueID, row.OrganizationID, row.Name, row.Status, row.SlotIntervalMinutes, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}
func (r *Repository) CreateCourt(ctx context.Context, organizationID, venueID, name string, interval int32) (model.Court, error) {
	venue, err := parseID(venueID)
	if err != nil {
		return model.Court{}, err
	}
	organization, err := parseID(organizationID)
	if err != nil {
		return model.Court{}, err
	}
	row, err := r.queries.CreateCourt(ctx, db.CreateCourtParams{ID: uuid.New(), VenueID: venue, OrganizationID: organization, Name: name, SlotIntervalMinutes: interval})
	if err != nil {
		return model.Court{}, mapWriteError("create court", err)
	}
	return courtModel(row.ID, row.VenueID, row.OrganizationID, row.Name, row.Status, row.SlotIntervalMinutes, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}
func (r *Repository) UpdateCourt(ctx context.Context, court model.Court, name string, interval int32, status string) (model.Court, error) {
	id, err := parseID(court.ID)
	if err != nil {
		return model.Court{}, err
	}
	venue, err := parseID(court.VenueID)
	if err != nil {
		return model.Court{}, err
	}
	organization, err := parseID(court.OrganizationID)
	if err != nil {
		return model.Court{}, err
	}
	row, err := r.queries.UpdateCourt(ctx, db.UpdateCourtParams{ID: id, VenueID: venue, OrganizationID: organization, Name: name, SlotIntervalMinutes: interval, Status: status})
	if err != nil {
		return model.Court{}, mapWriteError("update court", err)
	}
	return courtModel(row.ID, row.VenueID, row.OrganizationID, row.Name, row.Status, row.SlotIntervalMinutes, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}
func (r *Repository) ArchiveCourt(ctx context.Context, court model.Court) (model.Court, error) {
	id, err := parseID(court.ID)
	if err != nil {
		return model.Court{}, err
	}
	venue, err := parseID(court.VenueID)
	if err != nil {
		return model.Court{}, err
	}
	organization, err := parseID(court.OrganizationID)
	if err != nil {
		return model.Court{}, err
	}
	row, err := r.queries.ArchiveCourt(ctx, db.ArchiveCourtParams{ID: id, VenueID: venue, OrganizationID: organization})
	if err != nil {
		return model.Court{}, mapWriteError("archive court", err)
	}
	return courtModel(row.ID, row.VenueID, row.OrganizationID, row.Name, row.Status, row.SlotIntervalMinutes, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}
func parseID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, catalog.ErrInvalidInput
	}
	return id, nil
}
func mapWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "courts_venue_name_unique_idx" {
		return catalog.ErrCourtNameConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return catalog.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
func courtModel(id, venueID, organizationID uuid.UUID, name, status string, interval int32, createdAt, updatedAt time.Time) model.Court {
	return model.Court{ID: id.String(), VenueID: venueID.String(), OrganizationID: organizationID.String(), Name: name, Status: status, SlotIntervalMinutes: interval, CreatedAt: createdAt, UpdatedAt: updatedAt}
}

// IdentityCutoverReady blocks serving legacy data before verified transfer.
func (r *Repository) IdentityCutoverReady(ctx context.Context) (bool, error) {
	ready, err := r.queries.IdentityCutoverReady(ctx)
	if err != nil {
		return false, fmt.Errorf("read Identity cutover state: %w", err)
	}
	return ready, nil
}
