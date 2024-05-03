import {
  ApolloCache,
  ApolloClient,
  HttpLink,
  InMemoryCache,
  NormalizedCacheObject,
} from "@apollo/client";
import { createLogoutLink } from "../portal/apollo";
import { ViewerQueryDocument } from "../portal/query/viewerQuery.generated";

export function makeGraphQLEndpoint(graphqlOpaqueAppID: string): string {
  return `/api/apps/${encodeURIComponent(graphqlOpaqueAppID)}/graphql`;
}

export function makeClient(
  portalCache: ApolloCache<NormalizedCacheObject>,
  graphqlOpaqueAppID: string
): ApolloClient<unknown> {
  const httpLink = new HttpLink({
    uri: makeGraphQLEndpoint(graphqlOpaqueAppID),
  });
  const logoutLink = createLogoutLink(() => {
    portalCache.writeQuery({
      query: ViewerQueryDocument,
      data: {
        viewer: null,
      },
    });
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
