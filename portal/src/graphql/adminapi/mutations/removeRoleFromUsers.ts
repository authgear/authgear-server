import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  RemoveRoleFromUsersMutationDocument,
  RemoveRoleFromUsersMutationMutation,
} from "./removeRoleFromUsers.generated";

export function useRemoveRoleFromUsersMutation(): {
  removeRoleFromUsers: (roleKey: string, userIDs: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<RemoveRoleFromUsersMutationMutation>(
      RemoveRoleFromUsersMutationDocument
    );

  const removeRoleFromUsers = useCallback(
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

  return { removeRoleFromUsers, error, loading };
}
