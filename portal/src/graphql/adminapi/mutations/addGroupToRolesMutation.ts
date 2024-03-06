import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  AddGroupToRolesMutationDocument,
  AddGroupToRolesMutationMutation,
} from "./addGroupToRolesMutation.generated";

export function useAddGroupToRolesMutation(): {
  addGroupToRoles: (groupKey: string, roleKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<AddGroupToRolesMutationMutation>(
      AddGroupToRolesMutationDocument
    );

  const addGroupToRoles = useCallback(
    async (groupKey: string, roleKeys: string[]) => {
      await mutateFunction({
        variables: {
          groupKey,
          roleKeys,
        },
      });
    },
    [mutateFunction]
  );

  return { addGroupToRoles, error, loading };
}
