---
name: generate-schemas-and-gentype
description: Use when updating GraphQL schema or generated frontend types in authgear-server. Always regenerate checked-in schema/type artifacts with make export-schemas and portal npm run gentype instead of editing generated files by hand.
---

# Generate Schemas and Types

Use this skill when a change touches GraphQL schema files or generated frontend types.

## Workflow

1. Make the source change in Go or GraphQL documents first.
2. Regenerate backend schema artifacts with `make export-schemas`.
3. Only after the schema files have been updated, regenerate portal GraphQL types with `cd portal && npm run gentype`.
4. Verify the affected packages with the narrowest relevant tests or typecheck.

## Rules

- Do not hand-edit generated schema or type files when a generator exists.
- Keep the generated outputs in the same commit as the source change.
- Always run `make export-schemas` before `npm run gentype` for schema-related changes.
- If the generator rewrites more files than expected, inspect the diff before proceeding.

## Typical outputs

- `portal/src/graphql/adminapi/schema.graphql`
- `portal/src/graphql/adminapi/globalTypes.generated.ts`
- `portal/src/graphql/adminapi/query/*.generated.ts`
