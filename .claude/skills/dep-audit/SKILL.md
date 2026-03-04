---
name: dep-audit
description: Audit and fix dependency vulnerabilities in Go and Node.js packages. Runs govulncheck for Go and npm audit for each package.json directory. Commits fixes directory by directory.
argument-hint: "--fix"
---

Audit and fix dependency vulnerabilities in this project. Follow the steps below in order.

## Step 1: Go Vulnerability Check

Run from the project root and from `./k6` (it is a separate Go module with its own `go.mod`):

```
make govulncheck
cd k6 && make govulncheck
```

Parse the output:
- If there are **no vulnerabilities**, note it and move on.
- If there are **vulnerabilities**, for each affected module:
  1. Run `go list -m -u <module>` to find the latest available version.
  2. Compare the current version with the latest version:
     - If the major version changes (e.g. `v1.x.x` → `v2.x.x`), this is a **major version upgrade**. Generate a **Breaking Change Report** for it (see below) and ask the user to confirm before proceeding with that module.
     - If only minor/patch version changes, proceed automatically.
  3. After user confirmation (if needed), run `go get <module>@latest` then `go mod tidy` in all relevant directories:
     - `./`
     - `./custombuild`
     - `./e2e`
     - `./k6`

**Breaking Change Report for Go (major version bumps)** must include:
- Module name, current version → proposed version
- Link to the module's changelog or migration guide if available (check the module's repository)
- Known incompatibilities (import path changes, removed/renamed symbols)
- Ask: "Do you want to apply this major version upgrade? (yes/no)"

After updating Go deps:
- Run `make build` or `go build ./...` to verify the build.
- If the build breaks, report the compiler errors and ask the user how to proceed. Do not commit.
- If vulnerabilities cannot be fixed (no fix available), note them in an **Unfixable Issues Report** and notify the user.
- If fixes were applied and build passes, stage and commit: `git add go.sum go.mod custombuild/go.sum custombuild/go.mod e2e/go.sum e2e/go.mod k6/go.sum k6/go.mod` and commit with message: `chore: fix Go dependency vulnerabilities`

## Step 2: Node.js Audit — directory by directory

Process each directory that contains a `package.json` (excluding `node_modules`) in this order:
1. `portal/`
2. `authui/`
3. `portalgraphiql/`
4. `scripts/npm/`

For each directory:

1. `cd` into it and check for **stale overrides** in `package.json` before auditing. For each entry in the `"overrides"` field:
   - Identify what vulnerability or issue the override was originally added to fix (check git log: `git log --oneline -10 -- <dir>/package.json`).
   - Run `npm list <overridden-package> --all` to see the currently resolved versions.
   - Check if the parent package that originally required the vulnerable version has since released a patch that bundles a safe version on its own (i.e. the override is no longer needed to satisfy the advisory).
   - If the override is no longer needed, remove it, run `npm install`, and confirm with `npm audit` that there are no regressions before proceeding. Include the removal in the same commit as any other fixes for this directory.
2. `cd` into it and run `npm audit --json`.
2. Parse the output:
   - If **no vulnerabilities**, note it and move on to the next directory.
   - Before applying any fix, inspect what versions `npm audit fix` would install by running `npm audit fix --dry-run --json`. For each package that would be updated:
     - If the proposed fix is a **major version bump** (semver major increases), it is a potential breaking change regardless of whether the build will pass.
     - Collect all such packages into a **Breaking Change Report** listing:
       - Package name, current version → proposed version
       - Semver change type (major bump)
       - Any notes from the advisory about incompatible changes
     - Present the report and ask the user to confirm before applying those upgrades.
   - For fixes with only minor/patch version bumps, run `npm audit fix` automatically.
   - If the user confirmed breaking changes, run `npm audit fix --force` (only after confirmation).
   - If vulnerabilities are **unfixable via npm audit fix** (i.e. `npm audit fix --dry-run` shows no resolution), check if the vulnerability is in a **transitive dependency** whose parent has not yet released a patch:
     1. Identify the vulnerable transitive package and the minimum safe version that fixes it (from the advisory).
     2. Run `npm list <transitive-package> --all` to see **every installed version** of that package across the dependency tree. This is critical — there may be multiple versions installed at different semver ranges (e.g. `3.x`, `9.x`, `10.x`). Only the version(s) that fall in the advisory's vulnerable range need to be overridden.
     3. Identify the **direct parent package(s)** that pull in the vulnerable version (e.g. `eslint-plugin-sonarjs` → `minimatch@10.1.2`).
     4. Check the changelog/release notes between the currently-used vulnerable version and the safe version:
        - Look for any breaking changes (API removals, changed behavior, new peer-dep requirements).
        - If there are **no breaking changes**, add a **scoped override** to `package.json`, targeting only the parent package that pulls in the vulnerable dep. Do **NOT** use a flat global override — it will forcibly replace all other installed versions (including non-vulnerable ones) and break unrelated packages:
          ```json
          "overrides": {
            "<direct-parent-package>": {
              "<transitive-package>": "^<safe-version>"
            }
          }
          ```
          Then run `npm install` to apply, then `npm list <transitive-package> --all` to verify only the targeted instance changed, and `npm audit` to confirm the vulnerability is resolved.
        - If there **are breaking changes**, do not apply the override. Add the vulnerability to the **Unfixable Issues Report** and explain why the override is unsafe.
3. After applying fixes in a directory, verify the project still builds:
   - For `portal/`: `npm run build` (or `npm run typecheck` if faster)
   - For `authui/`: `npm run build`
   - For `portalgraphiql/`: `npm run build`
   - For `scripts/npm/`: skip build check (utility scripts)
4. Stage and commit changes for that directory immediately:
   - `git add <dir>/package.json <dir>/package-lock.json`
   - If overrides were added, include a comment in the commit body explaining which transitive dep was pinned, why, and a link to the advisory.
   - Commit message: `chore: fix npm dependency vulnerabilities in <dir>`

## Step 3: Final Summary

After all directories are processed, output a summary with three sections:

### Fixed
List every package that was updated (name, old version → new version, directory).

### Breaking Changes Applied
List any breaking changes that were confirmed and applied.

### Unfixable Issues
List any vulnerabilities that could not be resolved, including:
- Package name and version
- CVE/advisory ID
- Why it cannot be fixed — be specific:
  - No patch released yet by the package author
  - Transitive dependency where the override would introduce breaking changes (explain what breaks)
  - Locked by another dependency with no compatible version
- Recommended action (e.g., open issue with upstream, watch for future patch, manual code workaround)
