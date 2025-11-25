import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetDisabledStatusMutationMutation,
  SetDisabledStatusMutationDocument,
} from "./setDisabledStatusMutation.generated";

export function useSetDisabledStatusMutation(): {
  setDisabledStatus: (opts: {
    userID: string;
    isDisabled: boolean;
    reason: string | null;
    temporarilyDisabledFrom: Date | null;
    temporarilyDisabledUntil: Date | null;
  }) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<SetDisabledStatusMutationMutation>(
      SetDisabledStatusMutationDocument,
      {
        // Disabling a user will terminate all sessions.
        // So we have to refetch queries that fetch sessions.
        refetchQueries: ["UserQuery"],
      }
    );

  const setDisabledStatus = useCallback(
    async (opts: {
      userID: string;
      isDisabled: boolean;
      reason: string | null;
      temporarilyDisabledFrom: Date | null;
      temporarilyDisabledUntil: Date | null;
    }) => {
      const result = await mutationFunction({
        variables: {
          userID: opts.userID,
          isDisabled: opts.isDisabled,
          reason: opts.reason,
          temporarilyDisabledFrom: opts.temporarilyDisabledFrom,
          temporarilyDisabledUntil: opts.temporarilyDisabledUntil,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { setDisabledStatus, loading, error };
}
