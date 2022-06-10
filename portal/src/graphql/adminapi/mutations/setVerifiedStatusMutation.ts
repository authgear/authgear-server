import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetVerifiedStatusMutationMutation,
  SetVerifiedStatusMutationDocument,
} from "./setVerifiedStatusMutation.generated";

export function useSetVerifiedStatusMutation(userID: string): {
  setVerifiedStatus: (
    claimName: string,
    claimValue: string,
    isVerified: boolean
  ) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<SetVerifiedStatusMutationMutation>(
      SetVerifiedStatusMutationDocument
    );

  const setVerifiedStatus = useCallback(
    async (claimName: string, claimValue: string, isVerified: boolean) => {
      const result = await mutationFunction({
        variables: {
          userID,
          claimName,
          claimValue,
          isVerified,
        },
      });

      return !!result.data;
    },
    [mutationFunction, userID]
  );

  return { setVerifiedStatus, loading, error };
}
