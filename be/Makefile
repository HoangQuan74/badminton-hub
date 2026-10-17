SQLC_VERSION := v1.30.0
MIGRATE_VERSION := v4.19.1
IDENTITY_DIR := services/identity
USER_GATEWAY_DIR := gateways/user
ADMIN_GATEWAY_DIR := gateways/admin
SOCKET_GATEWAY_DIR := gateways/socket
PLATFORM_DIR := platform

PLATFORM_COVERAGE_MIN := 100

.PHONY: generate sqlc-check migrate-create migrate-up migrate-down migrate-version test coverage-platform vet build

generate:
	cd $(IDENTITY_DIR) && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

sqlc-check:
	cd $(IDENTITY_DIR) && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) vet

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_example" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) create -ext sql -dir internal/storage/postgres/schema/migrations -seq $(name)

migrate-up:
	@test -n "$(IDENTITY_DATABASE_URL)" || (echo "IDENTITY_DATABASE_URL is required" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) -path internal/storage/postgres/schema/migrations -database "$(IDENTITY_DATABASE_URL)" up

migrate-down:
	@test -n "$(IDENTITY_DATABASE_URL)" || (echo "IDENTITY_DATABASE_URL is required" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) -path internal/storage/postgres/schema/migrations -database "$(IDENTITY_DATABASE_URL)" down 1

migrate-version:
	@test -n "$(IDENTITY_DATABASE_URL)" || (echo "IDENTITY_DATABASE_URL is required" && exit 1)
	cd $(IDENTITY_DIR) && go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) -path internal/storage/postgres/schema/migrations -database "$(IDENTITY_DATABASE_URL)" version

test:
	for module in $(PLATFORM_DIR) services/* gateways/*; do \
		if [ -f "$$module/go.mod" ]; then (cd "$$module" && go test ./...); fi; \
	done

coverage-platform:
	@set -eu; \
	profile=$$(mktemp); \
	trap 'rm -f "$$profile"' EXIT; \
	(cd $(PLATFORM_DIR) && go test ./... -coverpkg=./... -coverprofile="$$profile"); \
	total=$$(cd $(PLATFORM_DIR) && go tool cover -func="$$profile" | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	awk -v total="$$total" -v minimum="$(PLATFORM_COVERAGE_MIN)" 'BEGIN { if (total < minimum) { printf "platform coverage %.1f%% is below required %.1f%%\n", total, minimum; exit 1 } printf "platform coverage %.1f%% meets required %.1f%%\n", total, minimum }'

vet:
	for module in $(PLATFORM_DIR) services/* gateways/*; do \
		if [ -f "$$module/go.mod" ]; then (cd "$$module" && go vet ./...); fi; \
	done

build:
	cd $(USER_GATEWAY_DIR) && go build ./cmd/user-gateway
	cd $(ADMIN_GATEWAY_DIR) && go build ./cmd/admin-gateway
	cd $(SOCKET_GATEWAY_DIR) && go build ./cmd/socket-gateway
	cd $(IDENTITY_DIR) && go build ./cmd/identity
