import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  ResetAccountLockoutMutationMutation,
  ResetAccountLockoutMutationDocument,
} from "./resetAccountLockoutMutation.generated";

export function useResetAccountLockoutMutation(): {
  resetAccountLockout: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<ResetAccountLockoutMutationMutation>(
      ResetAccountLockoutMutationDocument
    );

  const resetAccountLockout = useCallback(
    async (userID: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
        },
      });
      return !!result.data?.resetAccountLockout;
    },
    [mutationFunction]
  );
  return { resetAccountLockout, error, loading };
}
