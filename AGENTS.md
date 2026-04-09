# AGENTS.md

This repository is Authgear Server. Use the repo docs and skills below as the source of truth for agentic work.

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

- `api_design`
- `dep_audit`
- `new_siteadmin_api`
- `write_e2e_test`
- Repo-local skills for Go tests, Portal GraphQL operations, Go version updates, important-module updates, and vetted-position updates

## Verification

- Go changes: run `go test` on the affected package(s), and wider tests if the change is shared infrastructure.
- Frontend changes: run the relevant `npm run build` or `npm run typecheck` command in the affected package.
- Generated files: rerun the generator that owns the file, not a manual edit.
