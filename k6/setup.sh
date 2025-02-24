#!/bin/bash

set -eu

# Find out ./k6
directory_k6="$(set -eu; cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1; pwd)"

cd "$directory_k6"

docker compose down
docker buildx bake --allow=fs.read=.. --allow=fs.read=../postgres/postgres12 --file docker-compose.yaml --file docker-compose-build-cache.json
docker compose up -d db redis
(set -eux; sleep 3)

# Run migrations
docker compose run --rm authgear authgear database migrate up
docker compose run --rm authgear authgear audit database migrate up
docker compose run --rm authgear authgear images database migrate up
docker compose run --rm authgear-portal authgear-portal database migrate up

# Set up the project
docker compose run --rm -v "$directory_k6"/runtime:/tmp/authgear authgear-portal authgear-portal internal configsource create /tmp/authgear
docker compose run --rm authgear-portal authgear-portal internal domain create-custom loadtest --apex-domain localhost --domain localhost

# Start authgear
docker compose up -d authgear
(set -eux; sleep 1)
