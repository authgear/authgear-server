import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  DeleteAuthenticatorMutationMutation,
  DeleteAuthenticatorMutationDocument,
} from "./deleteAuthenticatorMutation.generated";

export function useDeleteAuthenticatorMutation(): {
  deleteAuthenticator: (authenticatorID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<DeleteAuthenticatorMutationMutation>(
      DeleteAuthenticatorMutationDocument
    );
  const deleteAuthenticator = useCallback(
    async (authenticatorID: string) => {
      const result = await mutationFunction({
        variables: {
          authenticatorID,
        },
      });

      return !!result.data?.deleteAuthenticator;
    },
    [mutationFunction]
  );

  return {
    deleteAuthenticator,
    loading,
    error,
  };
}
