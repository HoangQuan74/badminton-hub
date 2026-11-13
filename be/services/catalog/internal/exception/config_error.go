package exception

import "errors"

var (
	ErrDatabaseURLRequired = errors.New("CATALOG_DATABASE_URL is required")
)
