#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
compose_file="$repo_root/be/tests/integration/identity/compose.yml"
project_name="badminton-tenant-e2e-$$"
run_dir="$(mktemp -d "${TMPDIR:-/tmp}/badminton-tenant-e2e.XXXXXX")"
identity_pid=""

cleanup() {
  if [[ -n "$identity_pid" ]]; then
    kill "$identity_pid" 2>/dev/null || true
    wait "$identity_pid" 2>/dev/null || true
  fi
  docker compose -p "$project_name" -f "$compose_file" down --remove-orphans
}
trap cleanup EXIT

# Provision an isolated database with migrations and no development seeds.
docker compose -p "$project_name" -f "$compose_file" up -d --wait postgres
docker compose -p "$project_name" -f "$compose_file" run --rm migrate
db_address="$(docker compose -p "$project_name" -f "$compose_file" port postgres 5432)"
export TEST_IDENTITY_DATABASE_URL="postgres://onboard_test:onboard_test_local@$db_address/onboard_test?sslmode=disable"

# Verify persistence, authorization, concurrency, expiry, and rollback first.
cd "$repo_root/be/services/identity"
go test -race -count=1 ./...
go build -o "$run_dir/identity" ./cmd/identity
IDENTITY_DATABASE_URL="$TEST_IDENTITY_DATABASE_URL" \
IDENTITY_JWT_SECRET="onboard-e2e-only-secret-change-before-production" \
IDENTITY_GRPC_ADDR="127.0.0.1:0" \
  "$run_dir/identity" >"$run_dir/identity.log" 2>&1 &
identity_pid=$!

# Discover the OS-assigned listener without fixed-port collisions.
for attempt in {1..100}; do
  if ! kill -0 "$identity_pid" 2>/dev/null; then
    cat "$run_dir/identity.log"
    exit 1
  fi
  TEST_IDENTITY_GRPC_TARGET="$(python3 - "$run_dir/identity.log" <<'PY'
import json, sys
for line in open(sys.argv[1]):
    try:
        record = json.loads(line)
    except ValueError:
        continue
    if record.get('msg') == 'grpc server started':
        print(record['address'])
        break
PY
)"
  if [[ -n "$TEST_IDENTITY_GRPC_TARGET" ]]; then break; fi
  sleep 0.1
done
test -n "$TEST_IDENTITY_GRPC_TARGET"
export TEST_IDENTITY_GRPC_TARGET

# Run real HTTP adapters against the real Identity process and test database.
cd "$repo_root/be/gateways/user"
go test -race -count=1 ./...
printf 'Tenant CRUD HTTP -> gRPC -> PostgreSQL checks passed. Evidence directory: %s\n' "$run_dir"
