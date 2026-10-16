#!/usr/bin/env sh
set -eu

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
compose_file="$root_dir/devops/environments/db/compose.yml"
env_file="$root_dir/.env"

if [ ! -f "$env_file" ]; then
  echo "missing $env_file" >&2
  exit 1
fi

docker compose --project-name badminton-hub-db --env-file "$env_file" -f "$compose_file" exec -T postgres \
  sh -c '
    for database in "$POSTGRES_IDENTITY_DEV_DB" "$POSTGRES_IDENTITY_PROD_DB"; do
      exists=$(psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --tuples-only --no-align --command "SELECT 1 FROM pg_database WHERE datname = '\''$database'\''")
      if [ "$exists" != "1" ]; then
        createdb --username "$POSTGRES_USER" "$database"
        echo "created $database"
      else
        echo "exists $database"
      fi
    done
  '
