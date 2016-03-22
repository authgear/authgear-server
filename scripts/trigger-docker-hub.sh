#!/bin/bash -e

if [ -z "$DOCKER_HUB_TRIGGER_TOKEN" ]; then
    >&2 echo "DOCKER_HUB_TRIGGER_TOKEN is required."
    exit 1
fi

[ -n "$TRAVIS_TAG" ] && SOURCE_TAG="$TRAVIS_TAG"
[ -n "$TRAVIS_BRANCH" ] && SOURCE_BRANCH="$TRAVIS_BRANCH"
[ -n "$TRAVIS_REPO_SLUG" ] && SOURCE_REPO="$TRAVIS_REPO_SLUG"

DOCKER_HUB_REPO=skygeario/skygear-server
DOCKER_HUB_TRIGGER_URL=https://registry.hub.docker.com/u/$DOCKER_HUB_REPO/trigger/$DOCKER_HUB_TRIGGER_TOKEN/

if [ -n "$SOURCE_TAG" ]; then
    >&2 echo "Trigger build for tag $SOURCE_TAG on Docker Hub..."
    curl -H "Content-Type: application/json" --data '{"source_type": "Tag", "source_name": "'$SOURCE_TAG'"}' -X POST $DOCKER_HUB_TRIGGER_URL
elif [ -n "$SOURCE_BRANCH" ]; then
    >&2 echo "Trigger build for branch $SOURCE_BRANCH on Docker Hub..."
    curl -H "Content-Type: application/json" --data '{"source_type": "Branch", "source_name": "'$SOURCE_BRANCH'"}' -X POST $DOCKER_HUB_TRIGGER_URL
else
    >&2 echo "SOURCE_BRANCH is required."
    >exit 1
fi
