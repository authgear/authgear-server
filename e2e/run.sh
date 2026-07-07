#!/bin/bash -e

# Detect container runtime (docker or podman).
if command -v docker &>/dev/null; then
    CONTAINER_RUNTIME=docker
elif command -v podman &>/dev/null; then
    CONTAINER_RUNTIME=podman
else
    echo "Error: neither docker nor podman found." >&2
    exit 1
fi

# Kill any process listening on a TCP port.
# Tries lsof first (macOS / some Linux), then fuser (most Linux).
function kill_port {
    local port=$1
    if command -v lsof &>/dev/null; then
        kill -9 $(lsof -ti:"$port") > /dev/null 2>&1 || true
    elif command -v fuser &>/dev/null; then
        fuser -k "$port"/tcp > /dev/null 2>&1 || true
    fi
}

function setup {( set -e
    echo "[ ] Starting services..."
    $CONTAINER_RUNTIME compose up -d
    sleep 3

    echo "[ ] Building authgear..."
    make -C .. build BIN_NAME=dist/authgear TARGET=authgear
    make -C .. build BIN_NAME=dist/authgear-portal TARGET=portal
    export PATH=$PATH:../dist

    echo "[ ] Building e2e..."
    go build -o dist/e2e ./cmd/e2e
    go build -o dist/e2e-proxy ./cmd/proxy
    go build -o dist/e2e-smtp ./cmd/smtp
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

    echo "[ ] Starting e2e-smtp..."
    e2e-smtp > ./logs/e2e-smtp.log 2>&1 &
    success=false
    for i in $(seq 10); do \
        if [ "$(curl -sL -w '%{http_code}' -o /dev/null http://localhost:2525)" = "200" ]; then
            echo "    - started e2e-smtp."
            success=true
            break
        fi
        sleep 1
    done

    echo "[ ] DB migration..."
    authgear database migrate up
    authgear audit database migrate up
    authgear images database migrate up
    authgear-portal database migrate up

    echo "[ ] Starting siteadmin..."
    authgear-portal start siteadmin > ./logs/siteadmin.log 2>&1 &
    success=false
    for i in $(seq 10); do \
        if [ "$(curl -sL -w '%{http_code}' -o /dev/null http://localhost:4003/healthz)" = "200" ]; then
            echo "    - started siteadmin."
            success=true
            break
        fi
        sleep 1
    done
    if [ "$success" = false ]; then
        echo "Error: Failed to start siteadmin."
        exit 1
    fi
)}

function teardown {( set -e
    echo "[ ] Teardown..."
    kill_port 4000
    kill_port 4001
    kill_port 4002
    kill_port 4003
    kill_port 8080
    kill_port 2525
    $CONTAINER_RUNTIME compose down
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
