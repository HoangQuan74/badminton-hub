#!/usr/bin/env sh
set -eu

environment=${1:-}
if [ "$environment" != "dev" ] && [ "$environment" != "prod" ]; then
  echo "usage: rollback.sh <dev|prod>" >&2
  exit 2
fi

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
previous="$root_dir/.release/previous.env"
current="$root_dir/.release/current.env"

if [ ! -f "$previous" ]; then
  echo "no previous application release is recorded" >&2
  exit 1
fi

current_identity_image=
if [ -f "$current" ]; then
  current_identity_image=$(sed -n 's/^IDENTITY_IMAGE=//p' "$current" | tail -1)
fi
candidate_identity_image=
if [ -f "$root_dir/.release/candidate.env" ]; then
  candidate_identity_image=$(sed -n 's/^IDENTITY_IMAGE=//p' "$root_dir/.release/candidate.env" | tail -1)
fi

set -a
. "$previous"
set +a

# Releases created before the Identity extraction do not contain an
# IDENTITY_IMAGE entry. Derive the first compatible image name so rollback can
# still start the new Compose topology while using the previous gateway tag.
if [ -z "${IDENTITY_IMAGE:-}" ]; then
  IDENTITY_IMAGE=${current_identity_image:-${candidate_identity_image:-${USER_GATEWAY_IMAGE:-${API_IMAGE:-}}}}
  if [ -z "$IDENTITY_IMAGE" ]; then
    echo "previous release is missing IDENTITY_IMAGE and USER_GATEWAY_IMAGE" >&2
    exit 1
  fi
  export IDENTITY_IMAGE
fi

cp "$current" "$root_dir/.release/failed.env"
cp "$previous" "$current"
"$root_dir/devops/scripts/deploy.sh" "$environment"
echo "application rollback completed; database migrations were not reverted"
