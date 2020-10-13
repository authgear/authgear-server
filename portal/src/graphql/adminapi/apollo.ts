import { ApolloClient, InMemoryCache } from "@apollo/client";

export function makeClient(graphqlOpaqueAppID: string): ApolloClient<unknown> {
  const client = new ApolloClient({
    uri: `/api/apps/${encodeURIComponent(graphqlOpaqueAppID)}/graphql`,
    cache: new InMemoryCache({
      typePolicies: {
        User: {
          fields: {
            verifiedClaims: {
              // Take incoming data
              merge: false,
            },
          },
        },
      },
    }),
  });
  return client;
}
