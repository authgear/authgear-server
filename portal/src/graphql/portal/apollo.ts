import { ApolloClient, InMemoryCache } from "@apollo/client";

const client = new ApolloClient({
  uri: "/api/graphql",
  cache: new InMemoryCache({
    typePolicies: {
      App: {
        fields: {
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
