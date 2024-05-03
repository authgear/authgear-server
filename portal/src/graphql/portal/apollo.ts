import { createContext, useContext } from "react";
import {
  ApolloCache,
  ApolloClient,
  ApolloLink,
  HttpLink,
  InMemoryCache,
  NormalizedCacheObject,
} from "@apollo/client";
import { onError } from "@apollo/client/link/error";

export function createCache(): ApolloCache<NormalizedCacheObject> {
  return new InMemoryCache({
    typePolicies: {
      App: {
        fields: {
          resources: {
            merge: false,
          },
          resourceLocales: {
            merge: false,
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
          secretConfig: {
            // Take incoming data
            merge: false,
          },
        },
      },
      // NFTCollection does not have id, so we must teach Apollo what is the key of NFTCollection.
      NFTCollection: {
        keyFields: ["blockchain", "network", "contractAddress"],
      },
      // AppResource does not have id, so we must teach Apollo what is the key of AppResource.
      AppResource: {
        keyFields: ["path"],
      },
      // AppListItem doe snot have id, so we must teach Apollo what is the key of AppListItem.
      AppListItem: {
        keyFields: ["appID"],
      },
    },
  });
}

export function createLogoutLink(onLogout: () => void): ApolloLink {
  return onError(({ networkError, graphQLErrors }) => {
    const is401Error =
      networkError &&
      "statusCode" in networkError &&
      networkError.statusCode === 401;
    const isUnauthenticatedError = graphQLErrors?.some(
      (err) =>
        err.extensions.errorName === "Unauthorized" &&
        err.extensions.reason === "Unauthenticated"
    );
    if (is401Error || isUnauthenticatedError) {
      onLogout();
    }
  });
}

export function createClient(options: {
  cache: ApolloCache<NormalizedCacheObject>;
  onLogout: () => void;
}): ApolloClient<NormalizedCacheObject> {
  const { cache } = options;
  const httpLink = new HttpLink({ uri: "/api/graphql" });

  return new ApolloClient({
    link: createLogoutLink(options.onLogout).concat(httpLink),
    cache,
    name: "portal",
  });
}

const PortalClientContext = createContext<
  ApolloClient<NormalizedCacheObject> | undefined
>(undefined);

const PortalClientProvider = PortalClientContext.Provider;

export function usePortalClient(): ApolloClient<NormalizedCacheObject> {
  const client = useContext(PortalClientContext);
  if (client === undefined) {
    throw new Error("portal apollo client context provider not found");
  }
  return client;
}

export { PortalClientProvider };
