#!/usr/bin/env sh
set -eu

environment=${1:-}
root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
env_file="$root_dir/.env"

if [ ! -f "$env_file" ]; then
  echo "missing $env_file" >&2
  exit 1
fi

web_port=$(sed -n 's/^WEB_PORT=//p' "$env_file" | tail -1)
web_port=${web_port:-80}
url="http://127.0.0.1:$web_port"

attempt=1
while [ "$attempt" -le 30 ]; do
  if curl --fail --silent --show-error "$url/ready" | grep -q '"ready"'; then
    curl --fail --silent --show-error "$url/" >/dev/null
    echo "smoke test passed for ${environment:-environment} at $url"
    exit 0
  fi
  attempt=$((attempt + 1))
  sleep 2
done

echo "smoke test failed for ${environment:-environment} at $url" >&2
exit 1
