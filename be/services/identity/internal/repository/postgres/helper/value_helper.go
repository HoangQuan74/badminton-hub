package helper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// MapOptionalText preserves the difference between an omitted and an empty field.
func MapOptionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

func MapTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func MapOptionalTimestamp(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return MapTimestamp(*value)
}
