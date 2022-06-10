import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  UnscheduleAccountDeletionMutationMutation,
  UnscheduleAccountDeletionMutationDocument,
} from "./unscheduleAccountDeletionMutation.generated";

export function useUnscheduleAccountDeletionMutation(): {
  unscheduleAccountDeletion: (userID: string) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<UnscheduleAccountDeletionMutationMutation>(
      UnscheduleAccountDeletionMutationDocument,
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
