import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  RemoveUserFromRolesMutationDocument,
  RemoveUserFromRolesMutationMutation,
} from "./removeUserFromRoles.generated";

export function useRemoveUserFromRolesMutation(): {
  removeUserFromRoles: (userID: string, roleKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<RemoveUserFromRolesMutationMutation>(
      RemoveUserFromRolesMutationDocument
    );

  const removeUserFromRoles = useCallback(
    async (userID: string, roleKeys: string[]) => {
      await mutateFunction({
        variables: {
          userID,
          roleKeys,
        },
      });
    },
    [mutateFunction]
  );

  return { removeUserFromRoles, error, loading };
}
