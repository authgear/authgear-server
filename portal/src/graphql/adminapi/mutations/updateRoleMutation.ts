import { useCallback } from "react";
import {
  UpdateRoleMutationDocument,
  UpdateRoleMutationMutation,
} from "./updateRoleMutation.generated";
import { useMutation } from "@apollo/client";

export function useUpdateRoleMutation(): {
  updateRole: (role: {
    id: string;
    key?: string;
    name?: string | null;
    description?: string | null;
  }) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<UpdateRoleMutationMutation>(UpdateRoleMutationDocument);

  const updateRole = useCallback(
    async (role: {
      id: string;
      key?: string;
      name?: string | null;
      description?: string | null;
    }) => {
      await mutateFunction({
        variables: {
          id: role.id,
          key: role.key,
          name: role.name,
          description: role.description == null ? null : role.description,
        },
      });
    },
    [mutateFunction]
  );

  return { updateRole, error, loading };
}
