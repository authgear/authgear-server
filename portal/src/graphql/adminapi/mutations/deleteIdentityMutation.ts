import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  DeleteIdentityMutationMutation,
  DeleteIdentityMutationDocument,
} from "./deleteIdentityMutation.generated";

export function useDeleteIdentityMutation(): {
  deleteIdentity: (identityID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<DeleteIdentityMutationMutation>(DeleteIdentityMutationDocument);

  const deleteIdentity = useCallback(
    async (identityID: string) => {
      const result = await mutationFunction({
        variables: {
          identityID,
        },
      });
      return !!result.data?.deleteIdentity;
    },
    [mutationFunction]
  );
  return { deleteIdentity, error, loading };
}
