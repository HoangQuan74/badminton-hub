package identitytest

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/permission"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type permissionRepository struct {
	permission.PermissionRepository
	input permission.Input
	calls int
}

func (r *permissionRepository) CreatePermission(_ context.Context, input permission.Input) (model.Permission, error) {
	r.input = input
	r.calls++
	return model.Permission{ID: uuid.NewString(), Module: input.Module, Path: input.Path, Action: input.Action, Description: input.Description}, nil
}

func TestPermissionValidationAndNormalization(t *testing.T) {
	cases := []struct {
		name  string
		input permission.Input
		valid bool
	}{
		{"canonical values", permission.Input{Module: " Tenant ", Path: " /tenant ", Action: "read", Description: " View "}, true},
		{"blank module", permission.Input{Module: " ", Path: "/tenant", Action: "READ"}, false},
		{"blank path", permission.Input{Module: "tenant", Path: " ", Action: "READ"}, false},
		{"unknown action", permission.Input{Module: "tenant", Path: "/tenant", Action: "EXEC"}, false},
		{"nul byte", permission.Input{Module: "tenant", Path: "/tenant\x00", Action: "READ"}, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository := &permissionRepository{}
			service := permission.NewService(repository)
			_, err := service.CreatePermission(context.Background(), test.input)
			if !test.valid {
				if !errors.Is(err, identityException.ErrPermissionInvalidInput) || repository.calls != 0 {
					t.Fatalf("invalid values reached storage: calls=%d err=%v", repository.calls, err)
				}
				return
			}
			if err != nil || repository.input.Module != "tenant" || repository.input.Path != "/tenant" || repository.input.Action != "READ" || repository.input.Description != "View" {
				t.Fatalf("canonical permission: %+v err=%v", repository.input, err)
			}
		})
	}
}

func TestPermissionRejectsInvalidIdentifiersBeforeStorage(t *testing.T) {
	service := permission.NewService(&permissionRepository{})
	ctx := context.Background()
	if _, err := service.GetPermission(ctx, "invalid"); !errors.Is(err, identityException.ErrPermissionInvalidInput) {
		t.Fatal(err)
	}
	if _, err := service.UpdatePermission(ctx, "invalid", permission.Input{}); !errors.Is(err, identityException.ErrPermissionInvalidInput) {
		t.Fatal(err)
	}
	if err := service.DeletePermission(ctx, "invalid"); !errors.Is(err, identityException.ErrPermissionInvalidInput) {
		t.Fatal(err)
	}
}
