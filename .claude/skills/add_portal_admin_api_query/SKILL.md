---
name: add_portal_admin_api_query
description: Add a Portal Admin GraphQL query under portal/src/graphql/adminapi/query.
argument-hint: "<query name>"
---

Follow this skill when adding a Portal Admin GraphQL query.

## Workflow

1. Read `./portal/src/graphql/adminapi/schema.graphql` first.
2. Create the `.graphql` file under `./portal/src/graphql/adminapi/query`.
3. Include `cursor`, `pageInfo`, and `totalCount` when the query supports pagination and the user did not ask to omit them.
4. Query the fields of the entity itself, but avoid nested entities unless the user asked for them.
5. Run `cd ./portal && npm run gentype` after adding the file.

## Notes

- Keep the operation name consistent with existing Portal GraphQL naming.
- Do not invent schema fields; use only what exists in the schema file.
