import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  DeleteRoleMutationMutation,
  DeleteRoleMutationDocument,
} from "./deleteRoleMutation.generated";

export function useDeleteRoleMutation(): {
  deleteRole: (id: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<DeleteRoleMutationMutation>(DeleteRoleMutationDocument);

  const deleteRole = useCallback(
    async (id: string) => {
      const result = await mutationFunction({
        variables: {
          id,
        },
      });

      return !!result.data?.deleteRole.ok;
    },
    [mutationFunction]
  );

  return { deleteRole, loading, error };
}
