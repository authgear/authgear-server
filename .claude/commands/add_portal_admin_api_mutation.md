Provided the graphql schema: ./portal/src/graphql/adminapi/schema.graphql

Add .graphql a file under ./portal/src/graphql/adminapi/mutations for $ARGUMENTS.

- You should first read the graphql schema to understand how to write the mutation before you start.
- Query all fields inside the entity, except nested entities, unless the user explicitly mentioned they are required.
- After adding the graphql file, run `npm run gentype` inside ./portal to generate the ncessary code.
