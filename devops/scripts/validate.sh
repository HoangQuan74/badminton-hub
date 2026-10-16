#!/usr/bin/env sh
set -eu

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)

docker compose \
  --env-file "$root_dir/devops/environments/app/.env.example" \
  -f "$root_dir/devops/environments/app/compose.yml" \
  config --quiet

docker compose \
  --env-file "$root_dir/devops/environments/db/.env.example" \
  -f "$root_dir/devops/environments/db/compose.yml" \
  config --quiet

for script in "$root_dir"/devops/scripts/*.sh; do
  sh -n "$script"
done

echo "deployment configuration is valid"
