// Package grpcEndpoint provides the shared lifecycle for typed gRPC handlers.
package grpcEndpoint

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/platform/grpc/error"
)

// Run applies authorization, executes the typed operation, and maps domain
// errors to the gRPC status envelope owned by the caller's module.
func Run[Request any, Response any](
	ctx context.Context,
	request Request,
	mapper grpcError.ErrorMapper,
	authorize func(context.Context, Request) error,
	execute func(context.Context, Request) (Response, error),
) (Response, error) {
	var zero Response
	if authorize != nil {
		if err := authorize(ctx, request); err != nil {
			return zero, mapper.Map(err)
		}
	}
	response, err := execute(ctx, request)
	if err != nil {
		return zero, mapper.Map(err)
	}
	return response, nil
}

// Authorize adapts a token-based authorizer to generated request messages that
// expose GetAccessToken().
func Authorize[Request interface{ GetAccessToken() string }](
	authorizeToken func(context.Context, string) error,
) func(context.Context, Request) error {
	return func(ctx context.Context, request Request) error {
		return authorizeToken(ctx, request.GetAccessToken())
	}
}
