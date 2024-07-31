import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  MarkPasswordAsExpiredMutation,
  MarkPasswordAsExpiredDocument,
} from "./markPasswordAsExpired.generated";

export interface UseMarkPasswordAsExpiredMutationReturnType {
  markPasswordAsExpired: (userID: string, isExpired: boolean) => Promise<void>;
  loading: boolean;
  error: unknown;
}

export function useMarkPasswordAsExpiredMutation(): UseMarkPasswordAsExpiredMutationReturnType {
  const [mutationFunction, { error, loading }] =
    useMutation<MarkPasswordAsExpiredMutation>(MarkPasswordAsExpiredDocument);

  const markPasswordAsExpired = useCallback(
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
    markPasswordAsExpired,
    error,
    loading,
  };
}
