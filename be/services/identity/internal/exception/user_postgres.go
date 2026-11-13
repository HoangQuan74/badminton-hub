package exception

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsUserEmailConflict(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == pgerrcode.UniqueViolation && postgresError.ConstraintName == "users_email_unique_idx"
}
