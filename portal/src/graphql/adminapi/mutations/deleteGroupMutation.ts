import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  DeleteGroupMutationMutation,
  DeleteGroupMutationDocument,
} from "./deleteGroupMutation.generated";
import { GroupsListQueryDocument } from "../query/groupsListQuery.generated";

export function useDeleteGroupMutation(): {
  deleteGroup: (id: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<DeleteGroupMutationMutation>(DeleteGroupMutationDocument, {
      refetchQueries: [GroupsListQueryDocument],
    });

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
