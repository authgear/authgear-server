import { ApolloClient, InMemoryCache } from "@apollo/client";

export function makeClient(graphqlOpaqueAppID: string): ApolloClient<unknown> {
  const client = new ApolloClient({
    uri: `/api/apps/${encodeURIComponent(graphqlOpaqueAppID)}/graphql`,
    cache: new InMemoryCache(),
  });
  return client;
}
