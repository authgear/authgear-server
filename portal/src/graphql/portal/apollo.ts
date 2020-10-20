import { ApolloClient, InMemoryCache } from "@apollo/client";

interface AppResource {
  path: string;
}

const client = new ApolloClient({
  uri: "/api/graphql",
  cache: new InMemoryCache({
    typePolicies: {
      App: {
        fields: {
          resources: {
            keyArgs: [],
            merge(
              existing: AppResource[] | undefined,
              incoming: AppResource[]
            ) {
              const map = new Map<string, AppResource>();
              for (const r of existing ?? []) {
                map.set(r.path, r);
              }
              for (const r of incoming) {
                map.set(r.path, r);
              }
              return Array.from(map.values());
            },
          },
          domains: {
            // Take incoming data
            merge: false,
          },
          collaborators: {
            // Take incoming data
            merge: false,
          },
          collaboratorInvitations: {
            // Take incoming data
            merge: false,
          },
        },
      },
    },
  }),
});

export { client };
