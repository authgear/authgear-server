---
name: bootstrap-local-dev
description: Set up a fresh authgear-server development environment from scratch. Use when onboarding a new contributor, setting up a new machine, or when the user says "set up local dev from scratch" / "first-time setup". Covers the asdf + Homebrew install path on macOS; Nix users should follow CONTRIBUTING.md directly.
argument-hint: "(no args)"
---

Walk a contributor through first-time setup of this repo. Run steps in order, verify each one before moving on, and stop on the first failure so the user can fix it.

This skill assumes macOS + zsh/bash. Linux works for most steps but the Homebrew commands need adapting; if `uname` is not `Darwin`, warn the user up front and offer to fall back to CONTRIBUTING.md.

If the user already has a working local dev environment, do not run this skill — use the "Start local dev" recipe in `AGENTS.md` instead.

## Step 0: Preflight

1. Confirm the working directory is the repo root: `git rev-parse --show-toplevel` should print the path to `authgear-server`.
2. Read `.tool-versions` to know which versions are pinned. The current pin is:

   ```
   golang 1.26.2
   nodejs 20.19.5
   python 3.12.1
   ```

   If the file has different versions, use the file's values everywhere below.
3. Verify `asdf --version` ≥ `0.14`. If asdf is not installed, prompt the user to install it (https://asdf-vm.com/guide/getting-started.html) and pause — do not continue.

**Verify:** `git rev-parse --show-toplevel` succeeds, `.tool-versions` is readable, `asdf --version` prints a version.

## Step 1: Install pinned tool versions via asdf

Run:

```sh
asdf plugin add golang https://github.com/asdf-community/asdf-golang.git
asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git
asdf plugin add python
asdf install
```

Plugin-add commands are idempotent and will report "already added" if the plugin exists; that's fine, continue.

**Verify:** in the repo root, `asdf current` shows `golang`, `nodejs`, `python` all sourced from this repo's `.tool-versions` and marked installed. `which go` should point under `~/.asdf/`, and `go version` should match the pinned Go version.

## Step 2: Install Homebrew system dependencies

These need Homebrew prompts / sudo, so do NOT run them yourself. Prompt the user to run:

```sh
brew install make coreutils pkg-config icu4c vips libmagic
```

After they confirm completion, set up the required environment exports. Compute the right paths:

```sh
brew --prefix
brew --prefix icu4c
brew --prefix make
```

The user's shell must have these exports in effect for builds to work. Print the exact lines to add to `~/.zshrc` (or `~/.bash_profile`) using the resolved paths:

```sh
export PATH="$(brew --prefix make)/libexec/gnubin:$PATH"
export PKG_CONFIG_PATH="$(brew --prefix)/opt/icu4c/lib/pkgconfig"
export CGO_CFLAGS_ALLOW="-Xpreprocessor"
export CGO_CFLAGS="-I$(brew --prefix)/include"
export CGO_LDFLAGS="-L$(brew --prefix)/lib"
```

Prompt the user to add them and then `source` their shell rc (or open a new terminal). Do not proceed until they confirm.

**Verify:** `make --version` reports GNU Make 4.x or newer (not Apple's 3.81), and `pkg-config --modversion icu-i18n` succeeds.

## Step 3: Build vendored deps and frontends

Run `make vendor`. This installs golangci-lint, runs `go mod download`, `npm ci` in `scripts/npm`, `authui`, and `portal`, then builds both AuthUI and the portal frontend. Expect 5–15 minutes on a fresh clone.

**Verify:** the command exits 0, `resources/authgear/generated/` is non-empty, and `resources/portal/static/` is non-empty.

## Step 4: Environment file and config

1. `cp .env.example .env` (only if `.env` does not exist — do not overwrite).
2. Run the init command to generate `./var/`:

   ```sh
   go run ./cmd/authgear init \
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
   ```

3. Open `var/authgear.yaml`. Confirm the `oauth.clients` section contains BOTH the `portal` and `siteadmin` clients with the full redirect-URI list from CONTRIBUTING.md. If `init` produced a stub, replace the `oauth:` block with this:

   ```yaml
   oauth:
     clients:
     - client_id: portal
       issue_jwt_access_token: true
       name: Portal
       redirect_uris:
       - "http://portal.localhost:8000/oauth-redirect"
       - "http://portal.localhost:8001/oauth-redirect"
       - "http://portal.localhost:8010/oauth-redirect"
       - "http://portal.localhost:8011/oauth-redirect"
       - "com.authgear.example://host/path"
       - "com.authgear.example.rn://host/path"
       - com.authgear.exampleapp.flutter://host/path
       - com.authgear.exampleapp.xamarin://host/path
       post_logout_redirect_uris:
       - "http://portal.localhost:8000/"
       - "http://portal.localhost:8010/"
       x_application_type: traditional_webapp
     - client_id: siteadmin
       issue_jwt_access_token: true
       name: Site Admin
       post_logout_redirect_uris:
       - "http://localhost:8101"
       redirect_uris:
       - "http://localhost:8101/oauth2-redirect.html"
       x_application_type: spa
   ```

4. Check `/etc/hosts` contains:

   ```
   127.0.0.1 portal.localhost
   127.0.0.1 accounts.portal.localhost
   ```

   If missing, prompt the user to run (sudo is interactive — they must run it):

   ```sh
   sudo sh -c 'printf "\n127.0.0.1 portal.localhost\n127.0.0.1 accounts.portal.localhost\n" >> /etc/hosts'
   ```

**Verify:** `var/authgear.yaml` exists and the `oauth.clients` array has at least the two entries above. `getent hosts portal.localhost` (or `dscacheutil -q host -a name portal.localhost` on macOS) resolves to `127.0.0.1`.

## Step 5: Start Postgres

```sh
docker compose build postgres16
docker compose up -d postgres16 pgbouncer
```

**Verify:** `docker compose ps postgres16 pgbouncer` shows both as `running`/healthy. If pgbouncer keeps restarting, postgres is likely not ready yet — wait 5–10 seconds and re-check.

## Step 6: Apply database migrations

Run in this order:

```sh
go run ./cmd/authgear database migrate up
go run ./cmd/authgear audit database migrate up
go run ./cmd/authgear images database migrate up
go run ./cmd/authgear search database migrate up
go run ./cmd/portal database migrate up
```

Then create the initial portal config-source row in the DB:

```sh
go run ./cmd/portal internal configsource create ./var
```

**Verify:** re-run any one of the migrate commands; it should report no migrations to apply. `configsource create` should succeed without "duplicate key" — if it fails because the row already exists, that is fine, move on.

## Step 7: MinIO and buckets

```sh
docker compose up -d minio
```

The `mc` bucket setup requires an interactive shell inside the container with the env vars from `.env`. Prompt the user to run:

```sh
docker compose exec -it minio bash
# Inside the container:
mc alias set local http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
mc mb local/images
mc mb local/userexport
exit
```

**Verify:** ask the user to confirm both `mc mb` commands printed "Bucket created successfully".

## Step 8: Bring up everything else and start the servers

```sh
docker compose up -d
```

This starts the remaining services: redis, elasticsearch, minio (already up), the OTel collector, loki, tempo, prometheus, and the nginx proxy.

Then start the three long-running dev servers in the background, in three separate shells (use `run_in_background` for each), following the "Start local dev" recipe in `AGENTS.md`:

1. `make start` — main auth server.
2. `make start-portal` — portal backend.
3. `cd portal && npm start` — portal frontend (Vite).

**Verify:** poll each log file with an `until grep -q ... ; do sleep 2; done` loop until you see:

- `make start` log: a `serving request` or `open database` debug line.
- `make start-portal` log: an `open database` debug line.
- portal Vite log: `ready in <N> ms`.

If a server exits early, read its log file and report the error — do not auto-retry.

## Step 9: Create the bootstrap account and grant portal owner

With all servers running, create a user. Defaults match CONTRIBUTING.md (`user@example.com` / `password`); ask the user if they want to override before running.

```sh
go run ./cmd/authgear internal admin-api invoke \
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
  --variables-json "$(jq -cn --arg email "user@example.com" --arg password "password" '{email: $email, password: $password}')" | tee ./query_output
```

Then grant the user `owner` on the `accounts` project:

```sh
decoded_node_id="$(jq <./query_output --raw-output '.data.createUser.user.id' | basenc --base64url --decode)"
raw_id="${decoded_node_id#User:}"
go run ./cmd/portal internal collaborator add \
   --app-id accounts \
   --user-id "$raw_id" \
   --role owner
```

**Verify:** `query_output` contains a `data.createUser.user.id`. The `collaborator add` command exits 0.

## Step 10: Done

Print a summary:

- Portal: http://portal.localhost:8000
- Bootstrap credentials: the email/password from Step 9
- Background server task IDs (so the user can stop them later)

Then point the user at the post-setup topics in CONTRIBUTING.md when they need them:

- HTTPS via mkcert: `## Set up HTTPS to develop some specific features`
- Switch to DB config source: `## Switch to Database config source`
- LDAP local dev: `## Set up LDAP for local development`
- `sessionType=refresh_token` vs `sessionType=cookie`: `## Switching between sessionType=refresh_token and sessionType=cookie`

## Notes

- This skill is macOS + asdf only. For Nix Flakes users, follow CONTRIBUTING.md → "Install dependencies with Nix Flakes" instead.
- If any step fails, stop and report the error verbatim; do not skip ahead.
- Do not re-run `make vendor` for routine restarts — that step is for first-time setup only. To restart later, use the "Start local dev" recipe in `AGENTS.md`.
