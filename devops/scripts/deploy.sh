#!/usr/bin/env sh
set -eu

environment=${1:-}
if [ "$environment" != "dev" ] && [ "$environment" != "prod" ]; then
  echo "usage: deploy.sh <dev|prod>" >&2
  exit 2
fi

: "${USER_GATEWAY_IMAGE:?USER_GATEWAY_IMAGE is required}"
: "${IDENTITY_IMAGE:?IDENTITY_IMAGE is required}"
: "${WEB_IMAGE:?WEB_IMAGE is required}"
: "${IMAGE_TAG:?IMAGE_TAG is required}"

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
compose_file="$root_dir/devops/environments/app/compose.yml"
env_file="$root_dir/.env"
release_dir="$root_dir/.release"

if [ ! -f "$env_file" ]; then
  echo "missing $env_file; provision it from devops/environments/app/.env.example" >&2
  exit 1
fi

mkdir -p "$release_dir"
if [ -f "$release_dir/current.env" ]; then
  cp "$release_dir/current.env" "$release_dir/previous.env"
fi

cat > "$release_dir/candidate.env" <<EOF
USER_GATEWAY_IMAGE=$USER_GATEWAY_IMAGE
IDENTITY_IMAGE=$IDENTITY_IMAGE
WEB_IMAGE=$WEB_IMAGE
IMAGE_TAG=$IMAGE_TAG
EOF

cd "$root_dir"
export USER_GATEWAY_IMAGE IDENTITY_IMAGE WEB_IMAGE IMAGE_TAG

docker compose --project-name "badminton-hub-$environment" --env-file "$env_file" -f "$compose_file" pull
docker compose --project-name "badminton-hub-$environment" --env-file "$env_file" -f "$compose_file" run --rm migrate
docker compose --project-name "badminton-hub-$environment" --env-file "$env_file" -f "$compose_file" up -d --remove-orphans identity user-gateway web

"$root_dir/devops/scripts/smoke-test.sh" "$environment"
mv "$release_dir/candidate.env" "$release_dir/current.env"
echo "deployed $environment at image tag $IMAGE_TAG"
