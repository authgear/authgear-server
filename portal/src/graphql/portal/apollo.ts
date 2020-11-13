import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import { onError } from "@apollo/client/link/error";
import { authenticatedQuery } from "./query/authenticatedQuery";

interface AppResource {
  path: string;
  effectiveData?: string | null;
}

const cache = new InMemoryCache({
  typePolicies: {
    App: {
      fields: {
        resources: {
          merge(existing: AppResource[] | undefined, incoming: AppResource[]) {
            const map = new Map<string, AppResource>();
            // null for delete
            // undefined for query with path only
            for (const r of existing ?? []) {
              if (r.effectiveData !== undefined) {
                // merge only for data query
                map.set(r.path, r);
              }
            }
            for (const r of incoming) {
              if (r.effectiveData !== null) {
                map.set(r.path, r);
              } else {
                map.delete(r.path);
              }
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
