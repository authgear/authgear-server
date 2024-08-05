import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetPasswordExpiredMutation,
  SetPasswordExpiredDocument,
} from "./setPasswordExpired.generated";

export interface UseSetPasswordExpiredMutationReturnType {
  setPasswordExpired: (userID: string, isExpired: boolean) => Promise<void>;
  loading: boolean;
  error: unknown;
}

export function useSetPasswordExpiredMutation(): UseSetPasswordExpiredMutationReturnType {
  const [mutationFunction, { error, loading }] =
    useMutation<SetPasswordExpiredMutation>(SetPasswordExpiredDocument);

  const setPasswordExpired = useCallback(
    async (userID: string, isExpired: boolean) => {
      await mutationFunction({
        variables: {
          userID,
          isExpired,
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
