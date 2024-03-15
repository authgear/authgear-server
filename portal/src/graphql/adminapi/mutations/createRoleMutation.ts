import { useMutation } from "@apollo/client";
import {
  CreateRoleMutationDocument,
  CreateRoleMutationMutation,
} from "./createRoleMutation.generated";
import { useCallback } from "react";

export function useCreateRoleMutation(): {
  createRole: (role: {
    key: string;
    name?: string | null;
    description?: string | null;
  }) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<CreateRoleMutationMutation>(CreateRoleMutationDocument);

  const createRole = useCallback(
    async (role: {
      key: string;
      name?: string | null;
      description?: string | null;
    }) => {
      const result = await mutateFunction({
        variables: {
          key: role.key,
          name: role.name,
          description: role.description == null ? null : role.description,
        },
      });
      const roleID = result.data?.createRole.role.id ?? null;
      return roleID;
    },
    [mutateFunction]
  );

  return { createRole, error, loading };
}
