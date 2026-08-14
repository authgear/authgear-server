---
name: review-pr
description: Produce a structured code review report for the current branch's PR (purpose, API interface changes, code quality, bugs, security, performance), with a mandatory second verification pass. Use when the user asks "what does this PR do", "review this branch/PR", or asks to confirm/double-check review findings and look for more issues.
argument-hint: "[remote/branch to compare against, defaults to upstream main]"
---

Produce a code review report for the commits on the current branch that are not yet on the upstream main branch. This skill has three phases: scoping, an initial report, and a mandatory verification pass that re-derives every finding with an objective check (not re-reading the same code and nodding along).

## Phase 1: Scope the PR correctly

Do NOT assume the local `master`/`main` branch is up to date — in this repo it is frequently stale by hundreds of commits. Establish the true upstream base first:

1. Run `git remote -v` and identify the canonical upstream remote (the one pointing at `authgear/authgear-server`, conventionally named `authgear`; ask the user if ambiguous or absent).
2. `git fetch <upstream-remote> main --quiet`
3. `git merge-base HEAD <upstream-remote>/main` — this is the true base, regardless of what local `master` points at.
4. `git log --oneline <base>..HEAD` to enumerate exactly the commits in scope. Sanity-check this list against `git log --oneline -5` from the repo status — if the top commits don't match what the user expects, stop and re-check the remote/branch choice before proceeding.
5. `git diff --stat <upstream-remote>/main..HEAD` — if the stat includes files unrelated to the commit subjects (e.g. translation files, unrelated scripts), that means upstream `main` has advanced independently on other work; this is normal and not part of the PR. Don't let it inflate the reported scope — cross-check with `git show --stat <commit>` for each commit in step 4 individually to get the true per-commit file list.

## Phase 2: Write the report

Structure the report under these six headings. Base every claim on an actual diff read (`git show <commit>`, `git diff <base>..HEAD -- <path>`), not on commit message text alone — commit messages describe intent, not necessarily what the code does.

1. **Major purpose** — 3-6 sentences synthesizing what the set of commits accomplishes and why, grouped by theme if the commits span multiple concerns.

2. **API interface changes** — enumerate concretely:
   - GraphQL: new/changed types, fields, args in `schema.graphql` and the resolver file that backs them (`pkg/admin/graphql/`, `pkg/portal/graphql/`)
   - Go: new/changed exported struct fields, function signatures in `pkg/lib/`
   - HTTP/config: new endpoints, changed request/response shapes, `authgear.yaml` schema changes
   - Note whether each change is additive (safe) or a rename/removal (breaking)

3. **Code quality issues** — look specifically for:
   - Formatting/lint violations: run `gofmt -l <changed .go files>` and `gofmt -d` on any that fail; for frontend, run `npm run typecheck`/`eslint`/`stylelint`/`prettier` (full CI-parity check is mandatory in Phase 3, not optional)
   - i18n regressions: hardcoded user-facing strings in `portal/src` or `authui/src` that bypass `renderToString`/`FormattedMessage`/locale-data, including things like `Intl.DisplayNames`/`Intl.NumberFormat` calls hardcoded to a fixed locale instead of the active one, and non-JSX config objects (e.g. chart library dataset `label`/legend/tooltip config) that carry hardcoded English text. Exclude `.stories.tsx`/`.stories.ts` files from this check — Storybook demo props are dev-only and never shown to real users, so they don't need translation; hardcoded strings there are not a finding.
   - Dead/unreachable code introduced by the diff (e.g. a switch branch that can never be hit given the other cases)
   - Orphaned i18n keys: when the diff renames, consolidates, or drops the usage of a `FormattedMessage`/`renderToString` id, grep the old id across the whole frontend source tree (`portal/src`, `authui/src`), not just the changed file, to confirm zero remaining references — if none, the key should have been deleted from `locale-data/en.json` in the same diff and wasn't
   - Dead CSS rules: when the diff swaps or removes usages of CSS module classes (e.g. a component/layout replacement), grep the paired `.tsx`/`.ts` file(s) for each class still declared in the `.module.css` to confirm it has zero remaining usages — a leftover class from the old layout is a sign the migration didn't clean up its own CSS module
   - Regressions in refactors: a prop/constraint/validation that existed on one code path before the change and silently disappeared on a new code path introduced by the same diff (compare old branch vs new branch of an `if`/ternary the diff introduced)

4. **Bugs** — trace actual runtime behavior, don't just eyeball it. Prioritize:
   - Off-by-one and boundary math in date/time range logic
   - Label/copy vs. implementation mismatches (does the UI string still describe what the code now does after the diff?)
   - State persistence logic that can't distinguish "never set" from "explicitly set to empty/default" (common in localStorage-backed preference code: `if (parsed.length === 0) return defaults` silently overrides a deliberate "clear all" action)
   - Validation/constraints dropped when a component is swapped or a new branch is added (e.g. a `min`/`max` prop supported by the old component but not the new one it was replaced with in some branch)
   - **Silent truncation instead of explicit rejection**: when a request could ask for more than an endpoint can/should return (e.g. an oversized date range, an unbounded list), check whether the code silently caps the result (adding a `LIMIT`, slicing an array, etc.) versus validating the input upfront and rejecting it with a clear error when it exceeds what the endpoint supports. Silent truncation is a bug regardless of which subset it happens to keep — the caller has no signal that they asked for more than they got, which is especially dangerous for an "overview"/dashboard-style endpoint that then presents the truncated data as if it were complete. A fix that adds a `LIMIT` to address a prior "missing bounds"/unbounded-result-set finding should be treated as suspect by default — check whether the endpoint instead validates the request and returns an error for out-of-range input; don't accept "it's capped now, and the sort order looks right" as sufficient, since the caller-facing problem (a partial answer with no indication it's partial) isn't fixed by capping at all, correct direction or not.

   For each bug, state the concrete failure scenario (inputs/state → wrong output), not just "this looks suspicious."

5. **Security issues** — check every new/changed handler, resolver, and query for:
   - **Authorization/tenant scoping**: does every new GraphQL resolver, Admin API handler, or Site Admin handler enforce the same authz/role checks as sibling resolvers, and is every DB query scoped by `app_id`/tenant so one tenant cannot read or mutate another's data (IDOR)?
   - **Injection**: are all new SQL queries built through the existing query builder / parameterized placeholders (`db.SelectBuilder`, `?`/`$1` placeholders) with no raw string concatenation of user input, including inside JSONB path expressions (`data#>>'{...}'`) where a dynamic path segment could come from user input?
   - **Secrets/PII exposure**: does the diff log, return in an error message, or expose in a GraphQL field anything that should be redacted (tokens, full phone numbers/emails beyond what's already exposed, internal IDs that leak cross-tenant info)?
   - **SSRF/webhook risk**: if the diff adds or changes an outbound HTTP call (webhooks, image fetch, redirect URL), is the target validated/allow-listed the same way existing outbound calls are?
   - **XSS/output encoding**: for AuthUI template or portal-rendered HTML changes, is user-controlled data passed through the existing escaping helpers rather than raw-inserted?
   - **Input validation**: are new config fields, GraphQL args, and query params validated (length, enum membership, range) consistently with sibling fields, before use in a query or written to config?
   - Treat findings here as high-priority even if the report has few other findings — call out explicitly if a check produced no findings ("no authz regressions found in the resolvers touched by this diff") rather than omitting the section.

6. **Performance issues** — check for:
   - N+1 or repeated-round-trip patterns: multiple independent DB queries derived from the same base query/filter that could be one query or run concurrently, especially in hot-path screens (dashboards, list views)
   - Missing bounds: new queries/list endpoints without a `LIMIT`/pagination cap, or unbounded time ranges that can return arbitrarily large result sets
   - Missing indexes: new `WHERE`/`GROUP BY`/`ORDER BY` columns on large tables (`_audit_log` and other high-volume tables) that aren't covered by an existing index — check `pkg/lib/infra/db/migration/` (or `cmd/authgear/cmd/cmdaudit/migrations/`) for the relevant table's index list. An index existing is not sufficient: confirm the query's actual `WHERE`/`GROUP BY`/`ORDER BY` expression is textually identical to the indexed expression (same functions, same nesting/wrapping — e.g. `COALESCE(UPPER(x), '')` does **not** match an index built on `upper(x)`). Postgres only uses expression indexes on an exact syntactic match, not a semantically-equivalent one — cite the migration's exact expression next to the query's exact expression when making this claim.
   - Expensive DB-side aggregation with a cheaper alternative already in the diff: a `GROUP BY`/aggregate over a computed or JSONB-extracted expression on a high-volume table costs roughly one extraction+comparison per matching row, independent of indexing — an index only narrows which rows are scanned, it does not remove that per-row cost. Before accepting such a query as fine because "it's indexed," check whether a sibling code path in the same diff already derives equivalent data more cheaply (an in-process/cached lookup, a value already resolved elsewhere per-row) — if so, flag the inconsistency; the existence of the cheap path is evidence the DB aggregate is unnecessary, not just non-ideal.
   - Frontend: new state/effects that cause avoidable re-fetches or re-renders (e.g. a query re-running on every keystroke without debounce, a `useMemo`/`useCallback` dependency array that's broader than necessary), and large/blocking computations on the main thread for large datasets

## Phase 3: Verification pass (mandatory, do not skip)

Re-derive every reported bug, security issue, and quality issue marked as objective (formatting, build errors, missing index) using an independent check — do not just re-read the code a second time and confirm your own reasoning:

- **CI parity (mandatory, not "if easy to run")**: run the actual commands CI runs for every package the diff touches, not a subset, and quote the real output for each — passing typecheck alone does not mean CI passes. Match against `.github/workflows/run-checks.yaml`:
  - `portal/` touched: `cd portal && npm run typecheck && npm run eslint && npm run stylelint && npm run prettier && npm run test && npm run gentype && npm run build`
  - `authui/` touched: same sequence as above with `cd authui`
  - Go packages touched: `make lint`, `make test`, `make fmt`, `make check-tidy` from the repo root (equivalent to the `authgear-test` CI job)
  - If any command fails, fix it and re-run before reporting the PR as clean — do not report a fix as done, or hand back a report with no other findings, while a CI-equivalent command still fails locally. Quote the final passing output in the report.
- **Go formatting/build claims**: run `gofmt -l`/`gofmt -d` and `go build ./<affected packages>/...` (and `go vet` if suspicious) and quote the actual output.
- **Date/time/off-by-one bugs**: simulate the exact logic in isolation. For TypeScript/luxon logic, use `node -e` with the real library from `node_modules` (e.g. `require('luxon')` from the `portal/` directory) reproducing the exact function with representative inputs, and print the actual result — don't hand-derive it only in prose.
- **Silent-truncation claims**: for any range/list input newly bounded by a `LIMIT`, don't just confirm the clause exists — plug in the exact oversized input (e.g. a `rangeFrom`/`rangeTo` far wider than the cap) and state what actually happens: does the endpoint return an error (correct), or does it silently return a partial result with no indication to the caller that it's partial (a bug, regardless of which subset it kept)? Cite whether upfront input validation exists (a `Validate()`-style check rejecting out-of-range input) versus only a downstream `LIMIT`.
- **State-machine/persistence bugs**: simulate the save/load round trip with representative inputs (empty selection, partial selection, malformed data) in a small inline script and show the actual output for each case.
- **Prop/constraint-loss bugs**: grep the replacement component's prop types/interface to confirm the constraint genuinely has no equivalent (not just "wasn't passed in this call site") — cite the exact interface definition.
- **Orphaned i18n key / dead CSS claims**: grep the exact message id or CSS class name across the whole frontend source tree (not just the changed file) to confirm zero references remain, and quote the grep command and its (empty) output — don't flag a key/class as dead based on only checking the file(s) the diff touched.
- **Authz/tenant-scoping claims**: grep sibling resolvers/handlers for the authz call or `app_id` filter they use, and confirm the new code either has the matching call or is genuinely missing it — cite both the sibling's pattern and the new code's diff.
- **Injection claims**: cite the exact line constructing the query/path and confirm whether the interpolated value is attacker-controlled input or a fixed/internal constant.
- **Index/performance claims**: grep the migration files for the touched table to confirm whether an index covering the new query's filter/group columns exists or not — cite the migration file, don't guess. If an index exists, diff its exact indexed expression against the query's exact `WHERE`/`GROUP BY`/`ORDER BY` expression side by side (don't just confirm the column name appears in both) — a wrapper mismatch (extra `COALESCE`, missing `UPPER`, different JSONB path syntax) means the index can't actually be used. Separately, for any new DB aggregate on a high-volume table, grep the rest of the diff for a sibling code path resolving equivalent data a cheaper way (in-process lookup, cache, existing per-row field) — an index existing doesn't mean the aggregate itself was the right call.
- Re-check `git diff --stat` scope once more against `git show --stat` per commit to make sure no claim is actually about upstream `main`'s unrelated churn rather than this PR's diff.

While verifying, also spend one pass looking at any files that were touched by the PR but not yet fully read (check the per-commit `--stat` output from Phase 1 against what's actually been opened) — new bugs are often in the files that got the least attention in Phase 2.

## Output format

Present the final report as:
- The six Phase 2 headings, each finding stated concretely with a `file:line` reference. Include the Security and Performance headings even when empty — state explicitly that no issues were found rather than omitting the section.
- For a follow-up verification request specifically, present bugs/quality/security/performance issues as: original claim → verification method used → outcome (confirmed / refined / retracted), followed by any newly found issues from the extra pass.
- Keep prose tight — this is a report to act on, not an essay.
