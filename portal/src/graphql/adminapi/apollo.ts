import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import authgear from "@authgear/web";
import { createLogoutLink } from "../portal/apollo";

export function makeGraphQLEndpoint(graphqlOpaqueAppID: string): string {
  return `/api/apps/${encodeURIComponent(graphqlOpaqueAppID)}/graphql`;
}

export function makeClient(
  graphqlOpaqueAppID: string,
  onLogout: () => void
): ApolloClient<unknown> {
  // @ts-expect-error
  const fetch: typeof window.fetch = authgear.fetch.bind(authgear);

  const httpLink = new HttpLink({
    uri: makeGraphQLEndpoint(graphqlOpaqueAppID),
    fetch,
  });
  const logoutLink = createLogoutLink(() => {
    onLogout();
  });

  const client = new ApolloClient({
    link: logoutLink.concat(httpLink),
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
