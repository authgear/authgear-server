#!/usr/bin/env bash
# One-shot script to bootstrap a fresh authgear-server development environment.
# Safe to run multiple times (all steps are idempotent).
#
# Usage:
#   ./scripts/sh/setup-dev.sh
#
# Optional env vars:
#   ADMIN_EMAIL    – email for the initial admin account (default: user@example.com)
#   ADMIN_PASSWORD – password for the initial admin account (default: password)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# ── Colours ──────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
log()   { echo -e "${GREEN}[setup]${NC} $*"; }
warn()  { echo -e "${YELLOW}[setup]${NC} $*"; }
error() { echo -e "${RED}[setup]${NC} $*" >&2; }
die()   { error "$*"; exit 1; }

# ── Container tool (docker or podman) ────────────────────────────────────────
# Override by setting COMPOSE_CMD, e.g. COMPOSE_CMD="podman compose"
if [ -z "${COMPOSE_CMD:-}" ]; then
  if command -v docker &>/dev/null; then
    COMPOSE_CMD="docker compose"
  elif command -v podman &>/dev/null; then
    COMPOSE_CMD="podman compose"
  else
    die "Neither docker nor podman found. Please install one before running this script."
  fi
fi
log "Using compose command: $COMPOSE_CMD"

# ── Prerequisites ─────────────────────────────────────────────────────────────
for cmd in go jq; do
  command -v "$cmd" &>/dev/null || die "Required command not found: $cmd. Please install it before running this script."
done

GIT_HASH="git-$(git rev-parse --short=12 HEAD)"
GO_RUN=(go run -tags "authgeardev" -ldflags "-X github.com/authgear/authgear-server/pkg/version.Version=${GIT_HASH}")

# Disable OTEL and remote log shipping for all go run commands in this script.
# godotenv.Load() does not override existing env vars, so these take precedence
# over .env even after the process loads it.
export OTEL_METRICS_EXPORTER=none
export OTEL_TRACES_EXPORTER=none
export LOG_HANDLERS=console

# ── 1. .env ───────────────────────────────────────────────────────────────────
if [ ! -f .env ]; then
  log "Copying .env.example → .env"
  cp .env.example .env
else
  log ".env already exists, skipping"
fi

# ── 2. Container services ─────────────────────────────────────────────────────
log "Starting container services (postgres16, pgbouncer, redis, minio)..."
$COMPOSE_CMD up -d postgres16 pgbouncer redis minio

log "Waiting for PostgreSQL to be ready..."
for i in $(seq 1 60); do
  $COMPOSE_CMD exec -T postgres16 pg_isready -U postgres -q 2>/dev/null && break
  sleep 1
done
$COMPOSE_CMD exec -T postgres16 pg_isready -U postgres -q || die "PostgreSQL did not become ready after 60 s"

# ── 3. Generate config files ──────────────────────────────────────────────────
if [ ! -f ./var/authgear.yaml ]; then
  log "Generating config files in ./var ..."
  "${GO_RUN[@]}" ./cmd/authgear init \
    --interactive false \
    --output-folder ./var \
    --purpose portal \
    --app-id accounts \
    --public-origin 'http://accounts.portal.localhost:3100' \
    --portal-origin 'http://portal.localhost:8000' \
    --portal-client-id portal \
    --siteadmin-client-id siteadmin \
    --siteadmin-redirect-uri 'http://localhost:8101/oauth2-redirect.html' \
    --siteadmin-post-logout-redirect-uri 'http://localhost:8101' \
    --phone-otp-mode sms \
    --disable-email-verification true \
    --search-implementation postgresql
else
  log "Config files already present in ./var, skipping init"
fi

# ── 4. Database migrations ────────────────────────────────────────────────────
log "Running database migrations..."
"${GO_RUN[@]}" ./cmd/authgear database migrate up
"${GO_RUN[@]}" ./cmd/authgear audit database migrate up
"${GO_RUN[@]}" ./cmd/authgear images database migrate up
"${GO_RUN[@]}" ./cmd/authgear search database migrate up
"${GO_RUN[@]}" ./cmd/portal database migrate up

# ── 5. Portal config source ───────────────────────────────────────────────────
log "Creating portal config source (safe to re-run)..."
"${GO_RUN[@]}" ./cmd/portal internal configsource create ./var 2>&1 || true

# ── 6. MinIO buckets ──────────────────────────────────────────────────────────
log "Waiting for MinIO to be ready..."
for i in $(seq 1 60); do
  $COMPOSE_CMD exec -T minio mc alias set local http://localhost:9000 minio secretpassword &>/dev/null && break
  sleep 1
done
$COMPOSE_CMD exec -T minio mc alias set local http://localhost:9000 minio secretpassword \
  || die "MinIO did not become ready after 60 s"

log "Creating MinIO buckets (images, userexport)..."
$COMPOSE_CMD exec -T minio mc mb --ignore-existing local/images
$COMPOSE_CMD exec -T minio mc mb --ignore-existing local/userexport

# ── 7. Start authgear server temporarily to bootstrap the admin account ───────
log "Starting authgear server (needed to call the Admin API)..."
AUTHGEAR_LOG="$(mktemp)"
"${GO_RUN[@]}" ./cmd/authgear start >"$AUTHGEAR_LOG" 2>&1 &
AUTHGEAR_PID=$!
cleanup() {
  # `go run` spawns a child (the compiled binary). Killing just the `go run`
  # parent leaves the binary running. Kill children first, then the parent.
  if kill -0 "$AUTHGEAR_PID" 2>/dev/null; then
    log "Stopping temporary authgear server (pid $AUTHGEAR_PID)..."
    # Kill any child processes (the compiled authgear binary)
    for child in $(ps -o pid= --ppid "$AUTHGEAR_PID" 2>/dev/null); do
      kill "$child" 2>/dev/null || true
    done
    kill "$AUTHGEAR_PID" 2>/dev/null || true
    wait "$AUTHGEAR_PID" 2>/dev/null || true
  fi
  rm -f "${AUTHGEAR_LOG:-}"
}
trap cleanup EXIT

log "Waiting for the Admin API on port 3002..."
for i in $(seq 1 120); do
  (echo > /dev/tcp/localhost/3002) 2>/dev/null && break
  sleep 1
done
(echo > /dev/tcp/localhost/3002) 2>/dev/null || {
  error "Admin API (port 3002) did not become ready after 120 s. Server log:"
  cat "$AUTHGEAR_LOG" >&2
  exit 1
}

# ── 8. Create admin account ───────────────────────────────────────────────────
ADMIN_EMAIL="${ADMIN_EMAIL:-user@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-password}"

log "Creating admin account: $ADMIN_EMAIL"
QUERY_OUTPUT="$(mktemp)"

set +e
"${GO_RUN[@]}" ./cmd/authgear internal admin-api invoke \
  --app-id accounts \
  --endpoint "http://localhost:3002" \
  --host "accounts.portal.localhost:3100" \
  --query '
    mutation createUser($email: String!, $password: String!) {
      createUser(input: {
        definition: {
          loginID: {
            key: "email"
            value: $email
          }
        }
        password: $password
      }) {
        user {
          id
        }
      }
    }
  ' \
  --variables-json "$(jq -cn --arg email "$ADMIN_EMAIL" --arg password "$ADMIN_PASSWORD" '{email: $email, password: $password}')" \
  > "$QUERY_OUTPUT"
CREATE_EXIT=$?
set -e

if [ $CREATE_EXIT -ne 0 ]; then
  warn "Admin account creation returned an error (account may already exist). Raw output:"
  cat "$QUERY_OUTPUT" >&2
  rm -f "$QUERY_OUTPUT"
else
  cat "$QUERY_OUTPUT"

  # ── 9. Add account as owner of the accounts app ──────────────────────────────
  ENCODED_NODE_ID="$(jq -r '.data.createUser.user.id // empty' "$QUERY_OUTPUT")"
  rm -f "$QUERY_OUTPUT"

  if [ -z "$ENCODED_NODE_ID" ]; then
    warn "Could not extract user ID from the Admin API response. Skipping collaborator step."
  else
    # Decode base64url → "User:<uuid>", then strip the "User:" prefix.
    if command -v basenc &>/dev/null; then
      DECODED_NODE_ID="$(printf '%s' "$ENCODED_NODE_ID" | basenc --base64url --decode)"
    else
      DECODED_NODE_ID="$(python3 -c "
import base64, sys
s = sys.stdin.read().strip()
s += '=' * ((-len(s)) % 4)
print(base64.urlsafe_b64decode(s).decode())
" <<< "$ENCODED_NODE_ID")"
    fi
    RAW_ID="${DECODED_NODE_ID#User:}"

    log "Adding $ADMIN_EMAIL (id: $RAW_ID) as owner of the accounts app..."
    "${GO_RUN[@]}" ./cmd/portal internal collaborator add \
      --app-id accounts \
      --user-id "$RAW_ID" \
      --role owner
  fi
fi

log ""
log "Setup complete!"
log ""
log "Next steps:"
log "  1. Run 'make start'              – main auth server"
log "  2. Run 'make start-portal'       – portal backend"
log "  3. Run 'cd portal && npm start'  – portal frontend"
log "  4. Visit http://portal.localhost:8000 and log in as:"
log "       Email:    $ADMIN_EMAIL"
log "       Password: $ADMIN_PASSWORD"
log ""
log "Tip: edit /etc/hosts or configure dnsmasq so that"
log "  portal.localhost → 127.0.0.1"
log "  accounts.portal.localhost → 127.0.0.1"
