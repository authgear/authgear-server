#!/bin/bash -e

if [ -z "$DOCKER_HUB_AUTH_TRIGGER_URL" ]; then
    # DOCKER_HUB_AUTH_TRIGGER_URL is like: https://cloud.docker.com/api/build/v1/source/<uuid>/trigger/<uuid>/call/
    >&2 echo "DOCKER_HUB_AUTH_TRIGGER_URL is required."
    exit 1
fi

if [ -z "$DOCKER_HUB_GATEWAY_TRIGGER_URL" ]; then
    # DOCKER_HUB_GATEWAY_TRIGGER_URL is like: https://cloud.docker.com/api/build/v1/source/<uuid>/trigger/<uuid>/call/
    >&2 echo "DOCKER_HUB_GATEWAY_TRIGGER_URL is required."
    exit 1
fi

[ -n "$TRAVIS_TAG" ] && SOURCE_TAG="$TRAVIS_TAG"
[ -n "$TRAVIS_BRANCH" ] && SOURCE_BRANCH="$TRAVIS_BRANCH"
[ -n "$TRAVIS_REPO_SLUG" ] && SOURCE_REPO="$TRAVIS_REPO_SLUG"

declare -a TRIGGER_URLS=("$DOCKER_HUB_AUTH_TRIGGER_URL" "$DOCKER_HUB_GATEWAY_TRIGGER_URL")

function push_trigger() {
    for URL in "${TRIGGER_URLS[@]}"
    do
        curl -H "Content-Type: application/json" --data '{"source_type": "Tag", "source_name": "'$1'"}' -X POST $URL
    done
}

if [ -n "$SOURCE_TAG" ]; then
    >&2 echo "Trigger build for tag $SOURCE_TAG on Docker Hub..."
    push_trigger $SOURCE_TAG
elif [ -n "$SOURCE_BRANCH" ]; then
    >&2 echo "Trigger build for branch $SOURCE_BRANCH on Docker Hub..."
    push_trigger $SOURCE_BRANCH
else
    >&2 echo "SOURCE_BRANCH is required."
    exit 1
fi
