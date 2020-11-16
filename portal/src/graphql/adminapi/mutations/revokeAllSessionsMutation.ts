import { useCallback } from "react";
import { useMutation, gql } from "@apollo/client";
import { RevokeAllSessionsMutation } from "./__generated__/RevokeAllSessionsMutation";

const revokeAllSessionsMutation = gql`
  mutation RevokeAllSessionsMutation($userID: ID!) {
    revokeAllSessions(input: { userID: $userID }) {
      user {
        id
        sessions {
          edges {
            node {
              id
            }
          }
        }
      }
    }
  }
`;

export function useRevokeAllSessionsMutation(): {
  revokeAllSessions: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    RevokeAllSessionsMutation
  >(revokeAllSessionsMutation);

  const revokeAllSessions = useCallback(
    async (userID: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
        },
      });
      return !!result.data?.revokeAllSessions;
    },
    [mutationFunction]
  );
  return { revokeAllSessions, error, loading };
}
