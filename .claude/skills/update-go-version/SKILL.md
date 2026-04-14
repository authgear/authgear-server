---
name: update-go-version
description: Update the Go toolchain version across the repo.
argument-hint: "<new Go version>"
---

Use this skill when the Go version changes.

## Files to update

- `flake.nix`
- `.tool-versions`
- `go.mod`
- `custombuild/go.mod`
- `e2e/go.mod`
- `k6/go.mod`
- `packagetracker/go.mod.tpl`
- `custombuild/cmd/authgear/Dockerfile`
- `custombuild/cmd/portalx/Dockerfile`
- `cmd/portal/Dockerfile`
- `cmd/authgear/Dockerfile`
- `once/partial.dockerfile`

## Workflow

1. Update the version in every file above.
2. Use `nix store prefetch-file` to refresh the hash in `flake.nix`.
3. Run `make go-mod-tidy`.
4. Run `make once/Dockerfile`.
5. Verify the repo still builds if the version bump changed tool behavior.

## Notes

- Keep the version updates consistent across all Go modules.
- Do not leave one module on the old version; that creates hard-to-debug drift.
