import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  AddRoleToUsersMutationDocument,
  AddRoleToUsersMutationMutation,
} from "./addRoleToUsersMutation.generated";

export function useAddRoleToUsersMutation(): {
  addRoleToUsers: (roleKey: string, userIDs: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<AddRoleToUsersMutationMutation>(AddRoleToUsersMutationDocument);

  const addRoleToUsers = useCallback(
    async (roleKey: string, userIDs: string[]) => {
      await mutateFunction({
        variables: {
          roleKey,
          userIDs,
        },
      });
    },
    [mutateFunction]
  );

  return { addRoleToUsers, error, loading };
}
