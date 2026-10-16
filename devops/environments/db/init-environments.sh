#!/usr/bin/env sh
set -eu

create_database() {
  database=$1
  exists=$(psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --tuples-only --no-align --command "SELECT 1 FROM pg_database WHERE datname = '$database'")
  if [ "$exists" != "1" ]; then
    createdb --username "$POSTGRES_USER" "$database"
  fi
}

create_database "$POSTGRES_IDENTITY_DEV_DB"
create_database "$POSTGRES_IDENTITY_PROD_DB"
