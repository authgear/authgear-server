import { gql, QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import {
  UserQuery,
  UserQueryVariables,
  UserQuery_node_User,
} from "./__generated__/UserQuery";

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
        verifiedClaims {
          name
          value
        }
        standardAttributes
        sessions {
          edges {
            node {
              id
              type
              lastAccessedAt
              lastAccessedByIP
              displayName
            }
          }
        }
        isDisabled
        lastLoginAt
        createdAt
        updatedAt
      }
    }
  }
`;

interface UserQueryResult
  extends Pick<
    QueryResult<UserQuery, UserQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  user: UserQuery_node_User | null;
}

export function useUserQuery(userID: string): UserQueryResult {
  const { data, loading, error, refetch } = useQuery<
    UserQuery,
    UserQueryVariables
  >(userQuery, {
    variables: {
      userID,
    },
  });

  const user = useMemo(() => {
    return data?.node?.__typename === "User" ? data.node : null;
  }, [data]);

  return { user, loading, error, refetch };
}
