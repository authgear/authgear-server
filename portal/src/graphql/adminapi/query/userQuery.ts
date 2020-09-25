import { gql, QueryResult, useQuery } from "@apollo/client";
import { UserQuery, UserQueryVariables } from "./__generated__/UserQuery";

const userQuery = gql`
  query UserQuery($userID: ID!) {
    node(id: $userID) {
      __typename
      ... on User {
        id
        authenticators {
          edges {
            node {
              id
              type
              kind
              isDefault
              claims
              createdAt
              updatedAt
            }
          }
        }
        identities {
          edges {
            node {
              id
              type
              claims
              createdAt
              updatedAt
            }
          }
        }
        lastLoginAt
        createdAt
        updatedAt
      }
    }
  }
`;

export function useUserQuery(
  userID: string
): QueryResult<UserQuery, UserQueryVariables> {
  const userQueryResult = useQuery<UserQuery, UserQueryVariables>(userQuery, {
    variables: {
      userID,
    },
    fetchPolicy: "network-only",
  });

  return userQueryResult;
}
