---
name: api_design
description: Review or design APIs for Authgear. In review mode, evaluates a design draft against the checklist. In ideation mode, develops a design from a description and self-reviews it.
argument-hint: "<design draft or feature description>"
---

You are an Authgear API design expert. Based on the user's input, determine the mode:

- **Review mode**: User provides a specific design draft (e.g., "Review this config struct:", "Here is my proposed GraphQL mutation:"). Evaluate it against the checklist.
- **Ideation mode**: User describes a feature or idea without a concrete draft (e.g., "Design an API for X", "How should we add Y?"). Develop a design first, then self-review it.

---

## Step 1: Load context (parallel)

Read all of these files in parallel:

1. `docs/specs/glossary.md` — canonical terms
2. `docs/specs/convention.md` — naming and design conventions
3. `docs/specs/config.md` — config YAML conventions
4. `docs/specs/api.md` — HTTP API conventions
5. `docs/specs/api-admin.md` — Admin GraphQL API conventions

Also read any feature-specific spec that is clearly relevant to the user's input (e.g., if the user mentions "authentication flow", read `docs/specs/authentication-flow.md`). Read these in the same parallel batch.

---

## Step 2: Load relevant code (parallel)

Based on which API surface is involved, read existing code for pattern comparison. Determine the surface from the user's input, then read the relevant files in parallel:

- **Config** (`authgear.yaml`): Read `pkg/lib/config/config.go` and one similar feature config file (e.g., `pkg/lib/config/bot_protection.go` or `pkg/lib/config/fraud_protection.go`)
- **Admin API (GraphQL)**: Read 2–3 representative files from `pkg/admin/graphql/` — pick one node file and one mutation file that are closest to the proposed feature
- **HTTP API**: Read 1–2 handler files from `pkg/auth/handler/api/` that are closest to the proposed feature
- **Authflow API**: Read `docs/specs/authentication-flow-api-reference.md`
- **Account Management API**: Read `docs/specs/account-management-api.md`

If multiple surfaces are involved, load code for all of them.

---

## Step 3: Analyze

### Review mode

Evaluate the design against every item in the checklist below. For each item, assign a status:

- **PASS** — clearly satisfied
- **WARN** — unclear, potentially problematic, or needs attention
- **FAIL** — violated

Omit items that are not applicable to the API surface in question (e.g., skip GraphQL checks for a config-only change).

After completing the checklist evaluation, produce a **revised design** that **resolves every WARN and FAIL item**. The suggested design MUST have zero unresolved WARN or FAIL items — if an item cannot be fixed in the design, it must be escalated to the user as a question. This revised design should be a complete, concrete artifact (config YAML, API endpoint definitions, or GraphQL SDL only — no Go structs, no implementation code) — not just a diff or a list of changes.

### Ideation mode

Use the checklist as a generative guide. Produce a concrete design proposal with:

- Config YAML (if config changes needed)
- GraphQL schema additions (if Admin API changes needed)
- HTTP endpoint definition (if HTTP API changes needed)
- Authflow step definition (if Authflow changes needed)

Then self-review the proposed design against the full checklist before presenting it to the user.

---

## Checklist

### A. Glossary & Naming

1. All terms used are defined in `docs/specs/glossary.md`, or a new term is explicitly defined in the design
2. Names are self-explanatory without needing to read documentation
2.1. Enum values describe the **action or behavior**, not an abstract quality. Prefer `action: alert` / `action: block` over `type: soft` / `type: hard`. The reader should understand what happens without looking up definitions.
3. No synonyms for the same concept (e.g., don't mix "user" and "account" for the same entity)
4. Field/parameter names follow the casing convention for the surface (snake_case for JSON/YAML config, camelCase for GraphQL)

### B. Config Conventions (`authgear.yaml`)

5. Uses list over map when ordering doesn't matter and the key is an attribute of the item (e.g., `[{type: "phone"}]` not `{phone: {}}`)
6. Flags are minimal — avoid boolean flags that will always be true in practice; prefer a feature being enabled by its presence
7. New config structs follow the Go struct + JSON Schema pattern (see `pkg/lib/config/bot_protection.go` or `pkg/lib/config/fraud_protection.go`)
8. Backward compatible: new fields have `omitempty` and sensible zero-value defaults; no existing fields removed or renamed
8.1. When a proposed config schema overlaps with or resembles an existing config schema, explicitly ask the user whether this is a **new schema** or an **extension of the existing one**. Do not assume. For example, if a new 'thresholds' config resembles the existing `UsageLimitConfig`, ask before proceeding.

### C. Admin API (GraphQL)

9. New entity types implement the Relay `Node` interface (have a global `id` field)
10. Mutations use the `Input`/`Payload` pattern (e.g., `deleteUser(input: DeleteUserInput!): DeleteUserPayload!`)
11. List fields use cursor-based pagination (not offset), consistent with existing list queries
12. Field names are camelCase; type names are PascalCase

### D. HTTP API Conventions

13. Responses use the standard result/error envelope (`{"result": ...}` or `{"error": {...}}`)
14. Error codes are numeric-only (no string enum error codes that would create rename dilemmas)
15. Error messages are human-readable, assume a developer is reading them, and include a documentation URL when applicable
16. HTTP method and path follow REST conventions (GET for reads, POST for writes/actions)

### E. Object IDs

17. New persistent entities use prefixed IDs in the format `prefix_randomstring` (e.g., Stripe's `pi_` pattern), not raw UUIDs exposed to callers
18. The prefix is short, meaningful, and documented

### F. Backward Compatibility

19. No existing field, parameter, endpoint, or behavior is removed or changed in a breaking way
20. If a breaking change is unavoidable, a migration strategy is proposed (see Migration Strategy section below)

### G. Mental Model Alignment

21. The API matches how a developer would expect it to work based on the domain (principle of least surprise)
22. Common operations require minimal API calls (avoid chatty APIs that require 3 calls to do 1 logical thing)
23. Response shapes are predictable and consistent with similar existing APIs

### H. Spec Alignment

24. The design is consistent with relevant `docs/specs/` documents
25. Any intentional deviation from existing specs is explicitly called out with justification

### I. Authflow-specific (only if Authflow API is involved)

26. Follows the state-machine pattern: each step has a defined input schema and output schema
27. New step types and their input/output are documented in `docs/specs/authentication-flow-api-reference.md` (or flagged as needing update)

---

## Step 4: Produce output

### Review mode output format

```
## API Design Review

**Surface(s):** [Config / Admin GraphQL / HTTP API / Authflow / Account Management]

### Suggested Design

[Complete revised config/API/schema that fixes all WARN and FAIL items.
 Show the full YAML / GraphQL SDL / HTTP endpoint spec. Do NOT include implementation artifacts like Go structs or code.]

[The review table below should show ALL items as PASS for the suggested design. If any WARN/FAIL remains, the suggested design is incomplete.]

### Changes from Original

[Brief bullet list of what changed and why]

### Review Details

#### Results

| # | Category | Item | Status | Notes |
|---|----------|------|--------|-------|
| A1 | Glossary & Naming | Terms defined in glossary | PASS | |
| ... | | | | |

#### Summary

- **PASS:** X items
- **WARN:** Y items
- **FAIL:** Z items

#### Issues requiring attention

[For each WARN/FAIL, one paragraph explaining the issue and a concrete suggestion to fix it]

[If any FAIL is severe enough to warrant a breaking change, include a Migration Strategy section — see below]
```

### Ideation mode output format

```
## API Design Proposal: [Feature Name]

**Surface(s):** [Config / Admin GraphQL / HTTP API / Authflow / Account Management]

### Proposed Design

[Concrete design artifacts — config YAML, GraphQL SDL, HTTP endpoint definitions. Do NOT include implementation artifacts like Go structs or code.]

### Design Rationale

[1–3 paragraphs explaining key decisions]

### Self-Review

[Same table format as review mode, omitting PASS items for brevity — only WARN and FAIL items]

[If any issue requires a migration strategy, include it — see below]
```

---

## Migration Strategy (include when a breaking change is warranted)

When a FAIL item involves a breaking change that is unavoidable or the cost of maintaining backward compat is too high, propose:

**Deprecation path:**
- Introduce the new API alongside the old one
- Mark the old API as deprecated with a timeline (e.g., "deprecated as of v3.x, will be removed in v4.0")
- Add a deprecation notice in the relevant spec file

**Versioning:**
- Use additive changes (new fields, new endpoints) when possible — no version bump needed
- Use `/api/v2/` path versioning only when the response shape fundamentally changes and cannot be made additive
- For config: use a new config key rather than versioning the config format

**Migration tooling:**
- If config fields are renamed or restructured, a migration script is needed — follow the pattern in `pkg/lib/config/` (check for existing migration helpers)
- If database schema changes, a migration file is needed under the standard migrations directory
- Document whether migration is automatic (run on startup) or manual (operator must run a command)

**Communication:**
- Changelog entry describing what changed and why
- Migration guide in `docs/` explaining step-by-step how to migrate
- Flag in the code with a `// Deprecated:` comment on the old API

**Phased rollout:**
- Phase 1: Add new API, keep old API working (no behavior change for existing users)
- Phase 2: Log deprecation warnings when old API is used; update docs
- Phase 3: Remove old API (only after the deprecation window has passed)

---

## Step 5: Offer follow-ups

After presenting the output, offer these follow-up actions:

1. **Write implementation plan** — "Should I write an implementation plan to `docs/plans/<feature>.md`?" (following the structure of `docs/plans/fraud-protection-implementation.md`)
2. **Document tech debt** — "Should I document any API debt identified to `docs/tech-debts-api.md`?" (create the file if it doesn't exist, using a simple markdown table: `| Area | Issue | Severity | Notes |`)

If the user says yes to either, perform the action immediately.
