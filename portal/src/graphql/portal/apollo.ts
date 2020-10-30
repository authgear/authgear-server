import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import { onError } from "@apollo/client/link/error";
import { authenticatedQuery } from "./query/authenticatedQuery";

interface AppResource {
  path: string;
}

const cache = new InMemoryCache({
  typePolicies: {
    App: {
      fields: {
        resources: {
          merge(existing: AppResource[] | undefined, incoming: AppResource[]) {
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
});

const httpLink = new HttpLink({ uri: "/api/graphql" });

export const logoutLink = onError(({ networkError }) => {
  if (
    networkError &&
    "statusCode" in networkError &&
    networkError.statusCode === 401
  ) {
    cache.writeQuery({
      query: authenticatedQuery,
      data: {
        viewer: null,
      },
    });
  }
});

const client = new ApolloClient({
  link: logoutLink.concat(httpLink),
  cache,
});

export { client };
