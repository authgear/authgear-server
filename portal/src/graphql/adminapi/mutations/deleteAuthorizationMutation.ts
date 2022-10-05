import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  DeleteAuthorizationMutationDocument,
  DeleteAuthorizationMutationMutation,
} from "./deleteAuthorization.generated";

export function useDeleteAuthorizationMutation(): {
  deleteAuthorization: (authorizationID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<DeleteAuthorizationMutationMutation>(
      DeleteAuthorizationMutationDocument
    );

  const deleteAuthorization = useCallback(
    async (authorizationID: string) => {
      const result = await mutationFunction({
        variables: {
          authorizationID,
        },
      });
      return !!result.data?.deleteAuthorization;
    },
    [mutationFunction]
  );
  return { deleteAuthorization, error, loading };
}
