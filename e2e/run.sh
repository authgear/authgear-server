#!/bin/bash -e

. .env

function setup {
    echo "[ ] Starting services..."
    docker compose build authgear
    docker compose up -d
    sleep 5

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

    echo "[ ] DB migration..."
    docker compose exec authgear bash -c "
        authgear database migrate up
        authgear audit database migrate up
        authgear images database migrate up
    "
    docker compose exec portal bash -c "
        authgear-portal database migrate up
    "

    [ -d ./fixtures ] && for f in ./fixtures/*; do
        if [ -d "$f" ]; then
            echo "[ ] Creating project $f..."
            docker compose exec portal bash -c "
                authgear-portal internal configsource create $f \
                    --database-schema=\"$DATABASE_SCHEMA\" \
                    --database-url=\"$DATABASE_URL\"
            "
            docker compose exec authgear bash -c "
                authgear internal e2e import-users --config-source-dir=\"$f\"
            "
        fi
    done

    echo "[ ] Creating default domain..."
    docker compose exec portal bash -c "
        authgear-portal internal domain create-default \
            --database-schema=\"$DATABASE_SCHEMA\" \
            --database-url=\"$DATABASE_URL\" \
            --default-domain-suffix=\".portal.localhost\"
    "
}

function teardown {
    echo "[ ] Teardown..."
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
