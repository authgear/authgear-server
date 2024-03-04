import { useCallback } from "react";
import {
  UpdateRoleMutationDocument,
  UpdateRoleMutationMutation,
} from "./updateRoleMutation.generated";
import { useMutation } from "@apollo/client";

export function useUpdateRoleMutation(): {
  updateRole: (
    id: string,
    key?: string,
    name?: string,
    description?: string
  ) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<UpdateRoleMutationMutation>(UpdateRoleMutationDocument);

  const updateRole = useCallback(
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

  return { updateRole, error, loading };
}
