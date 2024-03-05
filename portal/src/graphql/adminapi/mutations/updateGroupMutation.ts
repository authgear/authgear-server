import { useCallback } from "react";
import {
  UpdateGroupMutationDocument,
  UpdateGroupMutationMutation,
} from "./updateGroupMutation.generated";
import { useMutation } from "@apollo/client";

export function useUpdateGroupMutation(): {
  updateGroup: (
    id: string,
    key?: string,
    name?: string,
    description?: string
  ) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<UpdateGroupMutationMutation>(UpdateGroupMutationDocument);

  const updateGroup = useCallback(
    async (id: string, key?: string, name?: string, description?: string) => {
      await mutateFunction({
        variables: {
          id,
          key,
          name,
          description,
        },
      });
    },
    [mutateFunction]
  );

  return { updateGroup, error, loading };
}
