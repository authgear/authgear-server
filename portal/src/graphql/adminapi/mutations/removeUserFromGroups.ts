import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  RemoveUserFromGroupsMutationDocument,
  RemoveUserFromGroupsMutationMutation,
} from "./removeUserFromGroups.generated";

export function useRemoveUserFromGroupsMutation(): {
  removeUserFromGroups: (userID: string, groupKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<RemoveUserFromGroupsMutationMutation>(
      RemoveUserFromGroupsMutationDocument
    );

  const removeUserFromGroups = useCallback(
    async (userID: string, groupKeys: string[]) => {
      await mutateFunction({
        variables: {
          userID,
          groupKeys,
        },
      });
    },
    [mutateFunction]
  );

  return { removeUserFromGroups, error, loading };
}
