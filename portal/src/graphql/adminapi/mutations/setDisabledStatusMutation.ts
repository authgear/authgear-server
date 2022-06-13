import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetDisabledStatusMutationMutation,
  SetDisabledStatusMutationDocument,
} from "./setDisabledStatusMutation.generated";

export function useSetDisabledStatusMutation(): {
  setDisabledStatus: (userID: string, isDisabled: boolean) => Promise<boolean>;
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
    async (userID: string, isDisabled: boolean) => {
      const result = await mutationFunction({
        variables: {
          userID,
          isDisabled,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { setDisabledStatus, loading, error };
}
