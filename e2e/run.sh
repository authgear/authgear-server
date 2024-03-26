#!/bin/bash -e

. .env

function setup {
    echo "[ ] Starting services..."
    docker compose up -d
    sleep 3

    echo "[ ] Building authgear..."
    make -C .. build BIN_NAME=dist/authgear TARGET=authgear
    make -C .. build BIN_NAME=dist/authgear-portal TARGET=portal

    echo "[ ] Starting authgear..."
    ../dist/authgear start > authgear.log 2>&1 &
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
    ../dist/authgear database migrate up
    ../dist/authgear audit database migrate up
    ../dist/authgear images database migrate up
    ../dist/authgear-portal database migrate up

    [ -d ./fixtures ] && for f in ./fixtures/*; do
        if [ -d "$f" ]; then
            echo "[ ] Creating project $f..."
            ../dist/authgear-portal internal configsource create $f \
                --database-schema="$DATABASE_SCHEMA" \
                --database-url="$DATABASE_URL"
            ../dist/authgear internal e2e import-users --config-source-dir="$f"
        fi
    done

    echo "[ ] Creating default domain..."
    ../dist/authgear-portal internal domain create-default \
        --database-schema="$DATABASE_SCHEMA" \
        --database-url="$DATABASE_URL" \
        --default-domain-suffix=".portal.localhost"
}

function teardown {
    echo "[ ] Teardown..."
    kill -9 $(lsof -ti:4000) > /dev/null 2>&1 || true
    docker compose down
}

function runtests {
    echo "[ ] Run tests..."
    go test ./tests/... -timeout 1m30s
}

function main {
    teardown || true
    setup
    runtests
    # teardown
}

BASEDIR=$(cd $(dirname "$0") && pwd)
PROJECTDIR=$(cd "${BASEDIR}/.." && pwd)
(
    cd "${BASEDIR}"
    main $@
)
