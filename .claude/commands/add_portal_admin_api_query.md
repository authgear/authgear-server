Provided the graphql schema: ./portal/src/graphql/adminapi/schema.graphql

Add a .graphql file under ./portal/src/graphql/adminapi/query for $ARGUMENTS.

- You should first read the graphql schema to understand how to write the query before you start.
- Include cursor, pageInfo, totalCount if available unless the user mentioned they are not needed.
- Query all fields inside the entity, except nested entities, unless the user explicitly mentioned they are required.
- After adding the graphql file, run `npm run gentype` inside ./portal to generate the ncessary code.
