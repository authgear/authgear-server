import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  AddUserToRolesMutationDocument,
  AddUserToRolesMutationMutation,
} from "./addUserToRolesMutation.generated";

export function useAddUserToRolesMutation(): {
  addUserToRoles: (userID: string, roleKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<AddUserToRolesMutationMutation>(AddUserToRolesMutationDocument);

  const addUserToRoles = useCallback(
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

  return { addUserToRoles, error, loading };
}
