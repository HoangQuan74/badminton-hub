#!/usr/bin/env bash

set -euo pipefail

cleanup() {
  kill "$user_pid" "$tenant_pid" "$admin_pid" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

npm run dev:user &
user_pid=$!
npm run dev:tenant &
tenant_pid=$!
npm run dev:admin &
admin_pid=$!

wait
