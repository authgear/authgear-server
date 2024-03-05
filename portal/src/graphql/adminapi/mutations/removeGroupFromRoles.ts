import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  RemoveGroupFromRolesMutationDocument,
  RemoveGroupFromRolesMutationMutation,
} from "./removeGroupFromRoles.generated";

export function useRemoveGroupFromRolesMutation(): {
  removeGroupFromRoles: (groupKey: string, roleKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<RemoveGroupFromRolesMutationMutation>(
      RemoveGroupFromRolesMutationDocument
    );

  const removeGroupFromRoles = useCallback(
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

  return { removeGroupFromRoles, error, loading };
}
