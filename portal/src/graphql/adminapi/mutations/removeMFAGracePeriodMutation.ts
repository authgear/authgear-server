import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  RemoveMfaGracePeriodMutationMutation,
  RemoveMfaGracePeriodMutationDocument,
} from "./removeMFAGracePeriod.generated";

export function useRemoveMFAGracePeriodMutation(): {
  removeMFAGracePeriod: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<RemoveMfaGracePeriodMutationMutation>(
      RemoveMfaGracePeriodMutationDocument
    );

  const removeMFAGracePeriod = useCallback(
    async (userID: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { removeMFAGracePeriod, loading, error };
}
