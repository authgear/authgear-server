import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  UnscheduleAccountAnonymizationMutationMutation,
  UnscheduleAccountAnonymizationMutationDocument,
} from "./unscheduleAccountAnonymizationMutation.generated";

export function useUnscheduleAccountAnonymizationMutation(): {
  unscheduleAccountAnonymization: (userID: string) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<UnscheduleAccountAnonymizationMutationMutation>(
      UnscheduleAccountAnonymizationMutationDocument,
      {
        // Disabling a user will terminate all sessions.
        // So we have to refetch queries that fetch sessions.
        refetchQueries: ["UserQuery"],
      }
    );

  const unscheduleAccountAnonymization = useCallback(
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
    unscheduleAccountAnonymization,
    loading,
    error,
  };
}
