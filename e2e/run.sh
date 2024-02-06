#!/bin/bash -e

. .env

function setup {
    echo "[ ] Starting services..."
    # docker compose build --pull
    docker compose up -d
    sleep 5

    echo "[ ] DB migration..."
    docker-compose exec authgear bash -c "
        authgear database migrate up
        authgear audit database migrate up
        authgear images database migrate up
        # portal database migrate up
    "

    # TODO(newman): Should use db fixture with CONFIG_SOURCE_TYPE=database

    echo "[ ] Starting authgear..."
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
}

function teardown {
    echo "[ ] Teardown..."
    docker compose down
}

function runtests {
    echo "[ ] Run tests..."
    # TODO(newman): Add tests
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
