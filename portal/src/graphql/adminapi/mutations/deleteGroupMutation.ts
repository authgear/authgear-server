import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  DeleteGroupMutationMutation,
  DeleteGroupMutationDocument,
} from "./deleteGroupMutation.generated";

export function useDeleteGroupMutation(): {
  deleteGroup: (id: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<DeleteGroupMutationMutation>(DeleteGroupMutationDocument);

  const deleteGroup = useCallback(
    async (id: string) => {
      const result = await mutationFunction({
        variables: {
          id,
        },
      });

      return !!result.data?.deleteGroup.ok;
    },
    [mutationFunction]
  );

  return { deleteGroup, loading, error };
}
