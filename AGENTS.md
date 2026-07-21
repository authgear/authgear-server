# AGENTS.md

This repository is Authgear Server. Use the repo docs and skills below as the source of truth for agentic work.

## Repository layout

```
authgear-server/
├── cmd/                        # Entry points (main packages)
│   ├── authgear/               # Core auth server binary
│   │   ├── main.go
│   │   ├── server/             # HTTP server wiring
│   │   ├── background/         # Background worker (cron jobs, queues)
│   │   ├── adminapi/           # Admin API server
│   │   ├── config/             # Config commands (validate, migrate)
│   │   └── images/             # Docker image utilities
│   ├── portal/                 # Portal backend binary
│   │   └── main.go
│   └── once/                   # One-shot migration / setup commands
│
├── pkg/                        # Shared Go packages
│   ├── lib/                    # Core business logic (auth flows, OAuth,
│   │   │                         sessions, SAML, OIDC, webhooks, etc.)
│   ├── auth/                   # AuthUI HTTP handlers & webapp routes
│   │   ├── handler/            # Request handlers
│   │   └── webapp/             # Server-side rendered web-app views
│   ├── admin/                  # Admin API GraphQL layer
│   │   └── graphql/
│   ├── portal/                 # Portal backend GraphQL layer
│   │   └── graphql/
│   ├── siteadmin/              # Site-admin API (global administration)
│   ├── api/                    # Shared API types / transport helpers
│   ├── resolver/               # Request resolver middleware
│   ├── util/                   # General-purpose utilities
│   └── images/                 # Image-processing helpers
│
├── authui/                     # AuthUI frontend (React/TypeScript)
│   │                             Compiled output is embedded into the binary.
│   └── src/
│
├── portal/                     # Portal frontend (React/TypeScript)
│   │                             The portal lets tenant admins manage their apps.
│   └── src/
│
├── resources/                  # Static resources embedded into binaries
│   ├── authgear/               # AuthUI templates, translations, assets
│   └── portal/                 # Portal static assets
│
├── e2e/                        # End-to-end test suite
│   └── tests/
│
├── docs/                       # Specs and design documents
│   ├── specs/                  # Feature and API specifications
│   └── api/                    # OpenAPI / schema definitions
│
├── hack/                       # Developer scripts and tooling
├── devtools/                   # Local development helpers
└── scripts/                    # CI / release scripts
```

**Key mappings:**

| What you want to change | Where to look |
|---|---|
| Auth UI (login/signup pages) | `authui/src/` (frontend), `pkg/auth/webapp/` (handlers), `resources/authgear/` (templates) |
| Portal (admin console UI) | `portal/src/` (frontend), `pkg/portal/graphql/` (backend) |
| Admin API (GraphQL) | `pkg/admin/graphql/`, `cmd/authgear/adminapi/` |
| Site Admin API | `pkg/siteadmin/`, `cmd/authgear/` |
| Core auth logic (OAuth, OIDC, sessions, flows) | `pkg/lib/` |
| Background jobs / workers | `cmd/authgear/background/` |
| Config schema & validation | `pkg/lib/config/`, `cmd/authgear/config/` |
| E2E tests | `e2e/tests/` |

## Documentation map

| Doc | Contents |
|---|---|
| [README.md](README.md) | Project overview, local setup, build, running Authgear |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution workflow, coding standards, testing, commit/PR process |
| [docs/specs/convention.md](docs/specs/convention.md) | Spec writing convention — required reading before authoring a new spec |
| [docs/specs/api.md](docs/specs/api.md) | Authgear public API spec (OAuth/OIDC, flows, endpoints) |
| [docs/specs/api-admin.md](docs/specs/api-admin.md) | Admin API spec (GraphQL, endpoints, auth) |
| [docs/specs/](docs/specs/) | Feature-specific specs — authoritative source for behavior rules |
| [docs/api/](docs/api/) | OpenAPI / schema definitions |
| [portal/docs/ARCHITECTURE.md](portal/docs/ARCHITECTURE.md) | Portal architecture: stack, GraphQL endpoints, providers, config/theming |
| [portal/docs/FRONTEND.md](portal/docs/FRONTEND.md) | Portal React SPA conventions: routing, GraphQL, styling, i18n, forms |
| [portal/docs/storybook.md](portal/docs/storybook.md) | Storybook conventions — read before adding or editing component stories |

## Common commands

Day-to-day shortcuts. See [CONTRIBUTING.md](CONTRIBUTING.md) for full setup.

### Start local dev

When the user says **"start local development environment"**, run these in order, each in its own terminal (commands 2–4 are long-running):

1. `docker compose up -d` — bring up Postgres, Redis, MinIO, etc.
2. `make start` — main auth server.
3. `make start-portal` — portal backend.
4. `cd portal && npm start` — portal frontend (Vite dev server).

Add `make authui-dev` as a fifth terminal only when editing AuthUI. Assumes the env is already set up per CONTRIBUTING.md.

### Other commands

| Command | What it does |
|---|---|
| `make vendor` | One-time bootstrap after cloning: installs golangci-lint, fetches Go deps, and builds the frontends. |
| `make test` | Run Go tests under `cmd/...` and `pkg/...`. |
| `make lint` / `make fmt` | Go lint / format. Frontend equivalents: `make -C portal lint`, `make -C portal fmt` (same for `authui`). |
| `cd portal && npm run typecheck` | Fast TypeScript-only check for the portal (same for `authui`). |
| `make generate` | Re-run `go generate` (wire, mocks, etc.). Run after changing generator sources. |
| `make export-schemas` | Regenerate config JSON schemas, GraphQL schemas, and portal `gentype`. |
| `make -C e2e run` | Run the e2e suite. |

For more targets, browse the root `Makefile` and `portal/package.json` / `authui/package.json`.

## Working rules

- Preserve user changes. Do not revert unrelated edits.
- Keep changes small and targeted.
- Prefer existing local patterns over inventing new ones.
- Reuse skills for repeatable workflows.
- If you change code, run the narrowest relevant test or build first, then broaden only if the change crosses package boundaries.

## Skills

Use existing repo skills instead of one-off instructions when they fit:

- `api-design`
- `bootstrap-local-dev` — **use this for first-time setup on a fresh machine** (asdf + Homebrew install, env files, DB migrations, MinIO, bootstrap account)
- `dep-audit`
- `new-siteadmin-api`
- `review-pr` — **mandatory before marking any code change complete** (see Verification below), also usable on demand for "review this PR/branch"
- `update-portal-ui` — **use this before adding or editing any portal UI page** (link components, i18n inline links, FluentUI Text pitfalls, hardcoded/untranslated text)
- `update-email-templates` — **use this before editing any email template, translation string, or email subject line** (`*.gotemplate`, `messages/translation.json`, `translation.json` subjects)
- `write-e2e-test` — **includes patterns for testing feature variants and actual functionality. Use this before writing, editing, or running e2e tests** (see "Common Patterns" section)
- Repo-local skills for Go tests, Portal GraphQL operations, Go version updates, important-module updates, and vetted-position updates

## Git workflow

### Branching

- Work on a feature branch, open a PR against `main`.
- Branch names are free-form, but keep them short and descriptive.

### Commit message style

Commits in this repo follow **imperative, present-tense** subject lines without a trailing period.
Use one of the established prefixes depending on the nature of the change:

| Situation | Format | Example |
|---|---|---|
| Feature in a named area | `[Area] Short description` | `[Site Admin] Collaborators API` |
| Bug fix | `Fix <what was wrong>` | `Fix RFC 5322 From header encoding for non-ASCII display names` |
| Routine maintenance | `chore: <what>` | `chore: Update deps` |
| Documentation | `doc: <what>` | `doc: Clarify that FetchUsageRecordsInRange toEndTime is exclusive` |
| Anything else | Plain imperative verb | `Add siteadmin UsageService for messaging and MAU usage` |

Common area tags seen in the codebase: `[Site Admin]`, `[Site Admin API]`, `[Portal]`, `[Migration Needed]`, `[Fraud Protection]`.

The **PR number is appended** to the subject of the squash-merge commit (added by GitHub, not manually): `[Site Admin] Usage API #5647`.

### Commit body

- Leave a blank line after the subject.
- If the work is tracked in Linear, add `ref DEV-XXXX` in the body.
- Keep individual (pre-merge) commits atomic and self-describing — the history shows a typical sequence: implementation plan → wiring / DI → service layer → handler stubs → real implementation → `.vettedpositions` update.

### Pull requests

- Push branches to your **personal fork**, not to `origin` (the upstream repo).
  - Add your fork as a remote if it isn't already: `git remote add <your-github-username> git@github.com:<your-github-username>/authgear-server.git`
  - Push with `git push <your-github-username> <branch-name>`.
  - Open PRs from `<your-github-username>:<branch-name>` targeting `authgear/authgear-server:main`.
- PRs are **squash-merged** into `main`.
- Keep the PR title identical to the intended squash-merge subject (without the `#N` suffix — GitHub adds that).
- Reference the Linear ticket in the PR description if one exists.

### Generated / bookkeeping files

- After line-number changes, run `make update-vettedpositions` (or the equivalent skill) rather than editing `.vettedpositions` by hand.
- Never edit generated files (`wire_gen.go`, GraphQL codegen output, translation JSON) manually — rerun their generators.

## Verification

- Go changes: run `go test` on the affected package(s), and wider tests if the change is shared infrastructure.
- Frontend changes: run the relevant `npm run build` or `npm run typecheck` command in the affected package.
- Generated files: rerun the generator that owns the file, not a manual edit.
- **Before marking any code change complete, run the `review-pr` skill on the diff and resolve every Bugs/Security/Performance/Code quality finding it reports** (or, if a finding is a false positive, say why). `review-pr` also runs the CI-equivalent lint/typecheck/test/build commands for every touched package — a change isn't complete until those pass, not just until the code "looks right." Don't rely on skills like `update-portal-ui` alone to catch issues: those encode specific known pitfalls, not a general correctness/performance/security review — `review-pr` is the general-purpose gate that catches what the narrower skills don't anticipate.
