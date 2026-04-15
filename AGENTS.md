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

## Read first

- `README.md`
- `CONTRIBUTING.md`
- `docs/specs/convention.md`
- `docs/specs/api.md`
- `docs/specs/api-admin.md`
- Any feature-specific spec under `docs/specs/`

## Working rules

- Preserve user changes. Do not revert unrelated edits.
- Keep changes small and targeted.
- Prefer existing local patterns over inventing new ones.
- Reuse skills for repeatable workflows.
- If you change code, run the narrowest relevant test or build first, then broaden only if the change crosses package boundaries.

## Skills

Use existing repo skills instead of one-off instructions when they fit:

- `api-design`
- `dep-audit`
- `new-siteadmin-api`
- `write-e2e-test`
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
