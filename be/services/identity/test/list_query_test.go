package identitytest

import (
	"testing"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

func TestNewListQuery(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		search    string
		from      string
		to        string
		wantError bool
		wantLimit int
	}{
		{
			name:      "defaults and trims search",
			search:    "  user@example.com  ",
			wantLimit: model.DefaultPageSize,
		},
		{
			name:      "accepts bounded page size",
			limit:     model.MaxPageSize,
			wantLimit: model.MaxPageSize,
		},
		{
			name:      "rejects oversized page",
			limit:     model.MaxPageSize + 1,
			wantError: true,
		},
		{
			name:      "rejects invalid date",
			from:      "2026-10-17",
			wantError: true,
		},
		{
			name:      "rejects reversed date range",
			from:      "2026-10-18T00:00:00Z",
			to:        "2026-10-17T00:00:00Z",
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := model.NewListQuery(
				test.limit,
				test.offset,
				test.search,
				test.from,
				test.to,
			)
			if (err != nil) != test.wantError {
				t.Fatalf("NewListQuery() error = %v, wantError %v", err, test.wantError)
			}
			if err != nil {
				return
			}
			if query.Limit != test.wantLimit {
				t.Fatalf("NewListQuery() limit = %d, want %d", query.Limit, test.wantLimit)
			}
			if test.name == "defaults and trims search" && query.Search != "user@example.com" {
				t.Fatalf("NewListQuery() search = %q, want trimmed search", query.Search)
			}
		})
	}

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.October, 31, 23, 59, 59, 0, time.UTC)
	query, err := model.NewListQuery(10, 20, "staff", from.Format(time.RFC3339), to.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("NewListQuery() valid date range error = %v", err)
	}
	if query.CreatedFrom == nil || query.CreatedTo == nil {
		t.Fatal("NewListQuery() did not parse the date range")
	}
}
