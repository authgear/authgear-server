import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetPasswordExpiredMutation,
  SetPasswordExpiredDocument,
} from "./setPasswordExpired.generated";

export interface UseSetPasswordExpiredMutationReturnType {
  setPasswordExpired: (userID: string, expired: boolean) => Promise<void>;
  loading: boolean;
  error: unknown;
}

export function useSetPasswordExpiredMutation(): UseSetPasswordExpiredMutationReturnType {
  const [mutationFunction, { error, loading }] =
    useMutation<SetPasswordExpiredMutation>(SetPasswordExpiredDocument);

  const setPasswordExpired = useCallback(
    async (userID: string, expired: boolean) => {
      await mutationFunction({
        variables: {
          userID,
          expired,
        },
      });
    },
    [mutationFunction]
  );

  return {
    setPasswordExpired,
    error,
    loading,
  };
}
