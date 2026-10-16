SQLC_VERSION := v1.30.0
MIGRATE_VERSION := v4.19.1
IDENTITY_DIR := services/identity
GATEWAY_DIR := services/gateway
PLATFORM_DIR := platform

.PHONY: generate sqlc-check migrate-create migrate-up migrate-down migrate-version test vet build

generate:
	cd $(IDENTITY_DIR) && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

sqlc-check:
	cd $(IDENTITY_DIR) && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) vet

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_example" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) create -ext sql -dir migrations -seq $(name)

migrate-up:
	@test -n "$(IDENTITY_DATABASE_URL)" || (echo "IDENTITY_DATABASE_URL is required" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) -path migrations -database "$(IDENTITY_DATABASE_URL)" up

migrate-down:
	@test -n "$(IDENTITY_DATABASE_URL)" || (echo "IDENTITY_DATABASE_URL is required" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) -path migrations -database "$(IDENTITY_DATABASE_URL)" down 1

migrate-version:
	@test -n "$(IDENTITY_DATABASE_URL)" || (echo "IDENTITY_DATABASE_URL is required" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) -path migrations -database "$(IDENTITY_DATABASE_URL)" version

test:
	for module in $(PLATFORM_DIR) services/*; do \
		if [ -f "$$module/go.mod" ]; then (cd "$$module" && go test ./...); fi; \
	done

vet:
	for module in $(PLATFORM_DIR) services/*; do \
		if [ -f "$$module/go.mod" ]; then (cd "$$module" && go vet ./...); fi; \
	done

build:
	cd $(GATEWAY_DIR) && go build ./cmd/gateway
	cd $(IDENTITY_DIR) && go build ./cmd/identity
