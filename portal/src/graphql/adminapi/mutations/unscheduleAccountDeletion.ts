import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { UnscheduleAccountDeletionMutation } from "./__generated__/UnscheduleAccountDeletionMutation";

const unscheduleAccountDeletionMutation = gql`
  mutation UnscheduleAccountDeletionMutation($userID: ID!) {
    unscheduleAccountDeletion(input: { userID: $userID }) {
      user {
        id
        isDisabled
        disableReason
        isDeactivated
        deleteAt
      }
    }
  }
`;

export function useUnscheduleAccountDeletionMutation(): {
  unscheduleAccountDeletion: (userID: string) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<UnscheduleAccountDeletionMutation>(
      unscheduleAccountDeletionMutation,
      {
        // Disabling a user will terminate all sessions.
        // So we have to refetch queries that fetch sessions.
        refetchQueries: ["UserQuery"],
      }
    );

  const unscheduleAccountDeletion = useCallback(
    async (userID: string) => {
      await mutationFunction({
        variables: {
          userID,
        },
      });
    },
    [mutationFunction]
  );

  return {
    unscheduleAccountDeletion,
    loading,
    error,
  };
}
