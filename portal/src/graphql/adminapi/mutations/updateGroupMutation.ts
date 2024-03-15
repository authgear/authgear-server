import { useCallback } from "react";
import {
  UpdateGroupMutationDocument,
  UpdateGroupMutationMutation,
} from "./updateGroupMutation.generated";
import { useMutation } from "@apollo/client";

export function useUpdateGroupMutation(): {
  updateGroup: (group: {
    id: string;
    key?: string;
    name?: string | null;
    description?: string | null;
  }) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<UpdateGroupMutationMutation>(UpdateGroupMutationDocument);

  const updateGroup = useCallback(
    async (group: {
      id: string;
      key?: string;
      name?: string | null;
      description?: string | null;
    }) => {
      await mutateFunction({
        variables: {
          id: group.id,
          key: group.key,
          name: group.name,
          description: group.description,
        },
      });
    },
    [mutateFunction]
  );

  return { updateGroup, error, loading };
}
