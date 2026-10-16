#!/usr/bin/env sh
set -eu

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
compose_file="$root_dir/devops/environments/db/compose.yml"
env_file="$root_dir/.env"
database_name=${1:-}

if [ -z "$database_name" ]; then
  echo "usage: $0 <service-database-name>" >&2
  exit 1
fi

backup_name="${database_name}_$(date -u +%Y%m%dT%H%M%SZ).sql.gz"

if [ ! -f "$env_file" ]; then
  echo "missing $env_file" >&2
  exit 1
fi

docker compose --project-name badminton-hub-db --env-file "$env_file" -f "$compose_file" exec -T postgres \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$1" | gzip' sh "$database_name" > "$root_dir/$backup_name"

echo "created $root_dir/$backup_name"
