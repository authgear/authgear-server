import { useMutation } from "@apollo/client";
import {
  CreateGroupMutationDocument,
  CreateGroupMutationMutation,
} from "./createGroupMutation.generated";
import { useCallback } from "react";

export function useCreateGroupMutation(): {
  createGroup: (group: {
    key: string;
    name?: string | null;
    description?: string | null;
  }) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<CreateGroupMutationMutation>(CreateGroupMutationDocument);

  const createGroup = useCallback(
    async (group: {
      key: string;
      name?: string | null;
      description?: string | null;
    }) => {
      const result = await mutateFunction({
        variables: {
          key: group.key,
          name: group.name ?? null,
          description: group.description ?? null,
        },
      });
      const groupID = result.data?.createGroup.group.id ?? null;
      return groupID;
    },
    [mutateFunction]
  );

  return { createGroup, error, loading };
}
