import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  RemoveGroupFromUsersMutationDocument,
  RemoveGroupFromUsersMutationMutation,
} from "./removeGroupFromUsers.generated";

export function useRemoveGroupFromUsersMutation(): {
  removeGroupFromUsers: (groupKey: string, userIDs: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<RemoveGroupFromUsersMutationMutation>(
      RemoveGroupFromUsersMutationDocument
    );

  const removeGroupFromUsers = useCallback(
    async (groupKey: string, userIDs: string[]) => {
      await mutateFunction({
        variables: {
          groupKey,
          userIDs,
        },
      });
    },
    [mutateFunction]
  );

  return { removeGroupFromUsers, error, loading };
}
