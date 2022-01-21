import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { ScheduleAccountDeletionMutation } from "./__generated__/ScheduleAccountDeletionMutation";

const scheduleAccountDeletionMutation = gql`
  mutation ScheduleAccountDeletionMutation($userID: ID!) {
    scheduleAccountDeletion(input: { userID: $userID }) {
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

export function useScheduleAccountDeletionMutation(): {
  scheduleAccountDeletion: (userID: string) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<ScheduleAccountDeletionMutation>(
      scheduleAccountDeletionMutation,
      {
        // Disabling a user will terminate all sessions.
        // So we have to refetch queries that fetch sessions.
        refetchQueries: ["UserQuery"],
      }
    );

  const scheduleAccountDeletion = useCallback(
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
    scheduleAccountDeletion,
    loading,
    error,
  };
}
