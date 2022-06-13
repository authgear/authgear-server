import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  ScheduleAccountDeletionMutationMutation,
  ScheduleAccountDeletionMutationDocument,
} from "./scheduleAccountDeletionMutation.generated";

export function useScheduleAccountDeletionMutation(): {
  scheduleAccountDeletion: (userID: string) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<ScheduleAccountDeletionMutationMutation>(
      ScheduleAccountDeletionMutationDocument,
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
