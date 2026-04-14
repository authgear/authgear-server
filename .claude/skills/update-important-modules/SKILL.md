---
name: update-important-modules
description: Refresh the repo's important Go modules when the dependency set drifts.
argument-hint: "<module names or audit target>"
---

Use this skill when `make ensure-important-modules-up-to-date` reports stale modules.

## Workflow

1. Run `make ensure-important-modules-up-to-date`.
2. If it reports outdated modules, update the affected module(s) in:
   - `./`
   - `./custombuild`
   - `./e2e`
3. Use `go get -u <module>` and then `go mod tidy` in each affected module.
4. Re-run `make ensure-important-modules-up-to-date` until it passes cleanly.

## Notes

- Keep the update scoped to the modules reported as stale.
- Do not update unrelated dependencies while fixing the important-module list.
