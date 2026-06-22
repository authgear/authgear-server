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

## Steps 4–9: One-shot setup

Run:

```sh
make setup
```

This runs `scripts/sh/setup-dev.sh`, which is idempotent and safe to re-run. It handles everything in one go:

1. Copies `.env.example` → `.env` (if not present)
2. Starts `postgres16`, `pgbouncer`, `redis`, `minio` via compose
3. Generates `./var/authgear.yaml` and `./var/authgear.secrets.yaml`
4. Runs all DB migrations (authgear, audit, images, search, portal)
5. Creates the portal config-source row
6. Creates MinIO buckets (`images`, `userexport`)
7. Creates the bootstrap admin account and grants it `owner` on the `accounts` app

**Prerequisites:** `docker` or `podman` on `$PATH`, and `jq`. Auto-detects `podman` if `docker` is not found; override with `COMPOSE_CMD="podman compose"`.

Default credentials are `user@example.com` / `password`. Override with:

```sh
ADMIN_EMAIL=dev@example.com ADMIN_PASSWORD=s3cr3t make setup
```

**Verify:** the script prints `Setup complete!` and the credentials. If it fails, read the error — the script stops on the first failure.

**Known gotcha — `go run` stray child process:** `go run` compiles to a temp binary child process. If the script is interrupted with `kill -9`, clean up manually: `pkill -f 'authgear start'`.

## Step 10: Done

Add DNS entries so the portal hostnames resolve (if not already present in `/etc/hosts`):

```
127.0.0.1 portal.localhost
127.0.0.1 accounts.portal.localhost
```

Then start the three long-running dev servers, each in its own terminal:

```sh
make start            # main auth server
make start-portal     # portal backend
cd portal && npm start  # portal frontend (Vite)
```

Print a summary:

- Portal: http://portal.localhost:8000
- Bootstrap credentials: printed at the end of `make setup`

Then point the user at the post-setup topics in CONTRIBUTING.md when they need them:

- HTTPS via mkcert: `## Set up HTTPS to develop some specific features`
- Switch to DB config source: `## Switch to Database config source`
- LDAP local dev: `## Set up LDAP for local development`
- `sessionType=refresh_token` vs `sessionType=cookie`: `## Switching between sessionType=refresh_token and sessionType=cookie`

## Notes

- This skill is macOS + asdf only. For Nix Flakes users, follow CONTRIBUTING.md → "Install dependencies with Nix Flakes" instead.
- If any step fails, stop and report the error verbatim; do not skip ahead.
- Do not re-run `make vendor` for routine restarts — that step is for first-time setup only. To restart later, use the "Start local dev" recipe in `AGENTS.md`.
