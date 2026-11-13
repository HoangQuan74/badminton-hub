package exception

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var (
	ErrPermissionInvalidInput = errors.New("permission.invalid_input")
	ErrPermissionNotFound     = errors.New("permission.not_found")
	ErrPermissionConflict     = errors.New("permission.already_exists")
)

var PermissionGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity permission operation failed",
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrStaffForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(ErrStaffNotFound, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrPermissionInvalidInput, codes.InvalidArgument, "invalid permission"),
	grpcError.MatchError(ErrPermissionNotFound, codes.NotFound, "permission not found"),
	grpcError.MatchError(ErrPermissionConflict, codes.AlreadyExists, "permission already exists"),
)

func PermissionPersistenceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPermissionNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ErrPermissionConflict
	}
	return err
}
