#!/bin/bash -e

function setup {
    echo "[ ] Starting services..."
    docker compose up -d
    sleep 3

    echo "[ ] Building authgear..."
    make -C .. build BIN_NAME=dist/authgear TARGET=authgear
    make -C .. build BIN_NAME=dist/authgear-portal TARGET=portal
    export PATH=$PATH:../dist

    echo "[ ] Building e2e..."
    make build
    export PATH=$PATH:./dist

    echo "[ ] Starting authgear..."
    authgear start > authgear.log 2>&1 &
    for i in $(seq 10); do \
        if [ "$(curl -sL -w '%{http_code}' -o /dev/null ${MAIN_LISTEN_ADDR}/healthz)" = "200" ]; then
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

    echo "[ ] DB migration..."
    authgear database migrate up
    authgear audit database migrate up
    authgear images database migrate up
    authgear-portal database migrate up
}

function teardown {
    echo "[ ] Teardown..."
    kill -9 $(lsof -ti:4000) > /dev/null 2>&1 || true
    docker compose down
}

function runtests {
    echo "[ ] Run tests..."
    go test ./... -timeout 1m30s
}

function main {
    teardown || true
    setup
    runtests
    teardown
}

BASEDIR=$(cd $(dirname "$0") && pwd)
PROJECTDIR=$(cd "${BASEDIR}/.." && pwd)
(
    cd "${BASEDIR}"
    main $@
)
