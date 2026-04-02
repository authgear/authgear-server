---
name: write-implementation-plan
description: Draft or update detailed implementation plans for authgear-server specs, design changes, and docs/plans files. Use when Codex needs to turn a spec or outdated plan into a concrete implementation plan with exact files, exact methods, runtime call flow, compatibility requirements, test coverage, and atomic commit steps.
---

# Write Implementation Plan

Write implementation plans as concrete engineering plans, not design notes.

## Workflow

1. Read the current spec, the existing plan, and the relevant code paths before writing the plan.
2. Identify the real integration points in code:
   - config types and schema
   - runtime entry points
   - event and delivery paths
   - storage and migration behavior
   - tests
3. Rewrite the plan around the current spec and the current codebase, not around stale assumptions from an older plan.

## Requirements

Write the plan with exact implementation intent.

- Name the exact files to create or modify.
- Name the exact methods, structs, and helpers to add or change.
- Describe the exact method call flow for important runtime paths.
- Separate config-layer types from runtime-layer types.
- Preserve existing codebase conventions for file placement and naming.
- Include backward-compatibility requirements explicitly.
- Include deployment and data-compatibility behavior explicitly when storage keys, payloads, or persisted state are involved.
- Include test coverage requirements.
- Include an atomic commit plan.

Do not write the plan at a hand-wavy level.

- Do not say “support this somehow”.
- Do not say “add helpers as needed”.
- Do not say “or equivalent new file”.
- Do not add helper methods that are not used by the plan’s call flow.
- Do not leave known behavior in an “open decisions” section.

## Required Detail

For runtime-heavy changes, include all of the following.

### Config Plan

- exact config structs
- exact package and file placement
- schema changes
- merge behavior
- migration behavior for deprecated config

### Runtime Plan

- exact entry points
- exact handler/service/limiter method signatures
- exact internal helper method signatures
- exact call sequence from request entry point to storage and delivery

### Storage and Migration Plan

- exact storage key format
- compatibility with existing keys or persisted data
- rollout behavior during deploy
- whether backfill, dual-read, or dual-write is needed

### Script Plan

If Redis/Lua/SQL scripts are involved, document:

- exact inputs
- exact outputs
- exact call sites
- exact success and failure behavior

### API Compatibility Plan

If an API error or payload changes, document:

- old fields to keep
- new fields to add
- which fields are legacy
- exact type and value compatibility rules

## Method Call Plans

For important logic, write explicit call plans.

1. Name the current entry point file and method.
2. State the new method that will be called.
3. State what that method resolves or computes.
4. State what helper it calls next.
5. State the return behavior on success and failure.

When multiple periods, thresholds, or branches exist, spell out exactly:

- what loops over what
- what is called once per request
- what is called once per period
- what is called once per configured limit

Remove ambiguity that could cause overcounting or wrong behavior.

## Atomic Commit Plan

Always include a final section with atomic commits.

Each commit entry must include:

- commit purpose
- exact files
- exact behavior or refactor scope
- whether generated files or wiring must be updated in the same commit

Keep commit boundaries reviewable and bisect-safe.

If dependency wiring changes, require generated wiring updates in the same commit.

## Authgear-Server Specific Rules

- keep config structs in `pkg/lib/config`
- place feature config structs in `feature_xx.go`
- separate feature-config structs from app-config structs even if their shapes are parallel
- keep runtime-only structs separate from config structs when runtime needs a unified resolved shape
- preserve backward compatibility for legacy config, Redis keys, API payloads, and error fields unless the spec explicitly removes it
- when the plan references existing Redis key names or legacy values, state exactly where they come from in current code
- if e2e coverage is needed, include a dedicated e2e commit and list the cases that must be covered

## Output Shape

Prefer this structure when it fits the task:

1. Goal / scope
2. Config model and schema
3. Runtime flow
4. Event / delivery flow
5. Compatibility and deployment behavior
6. File-level change plan
7. Test plan
8. Fixed behavioral decisions
9. Implementation order
10. Atomic commit plan

Adjust section names if needed, but keep the plan concrete.
