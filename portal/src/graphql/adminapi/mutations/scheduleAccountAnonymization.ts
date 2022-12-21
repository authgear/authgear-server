import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  ScheduleAccountAnonymizationMutationMutation,
  ScheduleAccountAnonymizationMutationDocument,
} from "./scheduleAccountAnonymizationMutation.generated";

export function useScheduleAccountAnonymizationMutation(): {
  scheduleAccountAnonymization: (userID: string) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<ScheduleAccountAnonymizationMutationMutation>(
      ScheduleAccountAnonymizationMutationDocument,
      {
        // Disabling a user will terminate all sessions.
        // So we have to refetch queries that fetch sessions.
        refetchQueries: ["UserQuery"],
      }
    );

  const scheduleAccountAnonymization = useCallback(
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
    scheduleAccountAnonymization,
    loading,
    error,
  };
}
