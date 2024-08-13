import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetMfaGracePeriodMutationMutation,
  SetMfaGracePeriodMutationDocument,
} from "./setMFAGracePeriod.generated";

export function useSetMFAGracePeriodMutation(): {
  setMFAGracePeriod: (userID: string, endAt: Date) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<SetMfaGracePeriodMutationMutation>(
      SetMfaGracePeriodMutationDocument
    );

  const setMFAGracePeriod = useCallback(
    async (userID: string, endAt: Date) => {
      const result = await mutationFunction({
        variables: {
          userID,
          endAt,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { setMFAGracePeriod, loading, error };
}
