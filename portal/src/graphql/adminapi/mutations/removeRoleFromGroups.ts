import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  RemoveRoleFromGroupsMutationDocument,
  RemoveRoleFromGroupsMutationMutation,
} from "./removeRoleFromGroups.generated";

export function useRemoveRoleFromGroupsMutation(): {
  removeRoleFromGroups: (roleKey: string, groupKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<RemoveRoleFromGroupsMutationMutation>(
      RemoveRoleFromGroupsMutationDocument
    );

  const removeRoleFromGroups = useCallback(
    async (roleKey: string, groupKeys: string[]) => {
      await mutateFunction({
        variables: {
          roleKey,
          groupKeys,
        },
      });
    },
    [mutateFunction]
  );

  return { removeRoleFromGroups, error, loading };
}
