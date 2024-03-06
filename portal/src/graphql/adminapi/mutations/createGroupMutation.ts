import { useMutation } from "@apollo/client";
import {
  CreateGroupMutationDocument,
  CreateGroupMutationMutation,
} from "./createGroupMutation.generated";
import { useCallback } from "react";

export function useCreateGroupMutation(): {
  createGroup: (
    key: string,
    name: string,
    description?: string
  ) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<CreateGroupMutationMutation>(CreateGroupMutationDocument);

  const createGroup = useCallback(
    async (key: string, name: string, description?: string) => {
      const result = await mutateFunction({
        variables: {
          key,
          name,
          description,
        },
      });
      const groupID = result.data?.createGroup.group.id ?? null;
      return groupID;
    },
    [mutateFunction]
  );

  return { createGroup, error, loading };
}
