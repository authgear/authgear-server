#!/bin/bash -e

function setup {( set -e
    echo "[ ] Starting services..."
    docker compose up -d
    sleep 3

    echo "[ ] Building authgear..."
    make -C .. build BIN_NAME=dist/authgear TARGET=authgear
    make -C .. build BIN_NAME=dist/authgear-portal TARGET=portal
    export PATH=$PATH:../dist

    echo "[ ] Building e2e..."
    go build -o dist/e2e ./cmd/e2e
    go build -o dist/e2e-proxy ./cmd/proxy
    export PATH=$PATH:./dist

    echo "[ ] Starting authgear..."
    authgear start > ./logs/authgear.log 2>&1 &
    success=false
    for i in $(seq 10); do \
        if [ "$(curl -sL -w '%{http_code}' -o /dev/null http://localhost:4000/healthz)" = "200" ]; then
            echo "    - started authgear."
            success=true
            break
        fi
        sleep 1
    done
    if [ "$success" = false ]; then
        echo "Error: Failed to start authgear."
        exit 1
    fi

    echo "[ ] Starting e2e-proxy..."
    e2e-proxy > ./logs/e2e-proxy.log 2>&1 &
    success=false
    for i in $(seq 10); do \
        if [ "$(curl -sL -w '%{http_code}' -o /dev/null http://localhost:8080)" = "400" ]; then
            echo "    - started e2e-proxy."
            success=true
            break
        fi
        sleep 1
    done
    if [ "$success" = false ]; then
        echo "Error: Failed to start e2e-proxy."
        exit 1
    fi

    echo "[ ] DB migration..."
    authgear database migrate up
    authgear audit database migrate up
    authgear images database migrate up
    authgear-portal database migrate up
)}

function teardown {( set -e
    echo "[ ] Teardown..."
    kill -9 $(lsof -ti:4000) > /dev/null 2>&1 || true
    kill -9 $(lsof -ti:8080) > /dev/null 2>&1 || true
    docker compose down
)}

function tests {( set -e
    echo "[ ] Run tests..."
    # Use -count 1 to disable cache. We want to run the tests without caching.
    go test ./... -count 1 -v -timeout 10m -parallel 5
)}

function main {( set -e
    teardown || true
    setup
    trap "teardown || true" EXIT
    tests
)}

BASEDIR=$(cd $(dirname "$0") && pwd)
PROJECTDIR=$(cd "${BASEDIR}/.." && pwd)
(
    set -e
    cd "${BASEDIR}"
    if [ "$1" ]; then
        $1
    else
        main $@
    fi
)
