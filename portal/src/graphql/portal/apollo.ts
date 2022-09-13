import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import { onError } from "@apollo/client/link/error";
import { ViewerQueryDocument } from "./query/viewerQuery.generated";

const cache = new InMemoryCache({
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
      query: ViewerQueryDocument,
      data: {
        viewer: null,
      },
    });
  }
});

const client = new ApolloClient({
  link: logoutLink.concat(httpLink),
  cache,
  name: "portal",
});

export { client };
