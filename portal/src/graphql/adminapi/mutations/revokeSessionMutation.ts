import { useCallback } from "react";
import { useMutation, gql } from "@apollo/client";
import { RevokeSessionMutation } from "./__generated__/RevokeSessionMutation";

const revokeSessionMutation = gql`
  mutation RevokeSessionMutation($sessionID: ID!) {
    revokeSession(input: { sessionID: $sessionID }) {
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

export function useRevokeSessionMutation(): {
  revokeSession: (sessionID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<RevokeSessionMutation>(revokeSessionMutation);

  const revokeSession = useCallback(
    async (sessionID: string) => {
      const result = await mutationFunction({
        variables: {
          sessionID,
        },
      });
      return !!result.data?.revokeSession;
    },
    [mutationFunction]
  );
  return { revokeSession, error, loading };
}
