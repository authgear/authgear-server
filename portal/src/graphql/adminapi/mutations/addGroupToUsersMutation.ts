import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  AddGroupToUsersMutationDocument,
  AddGroupToUsersMutationMutation,
} from "./addGroupToUsersMutation.generated";

export function useAddGroupToUsersMutation(): {
  addGroupToUsers: (groupKey: string, userIDs: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<AddGroupToUsersMutationMutation>(
      AddGroupToUsersMutationDocument
    );

  const addGroupToUsers = useCallback(
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

  return { addGroupToUsers, error, loading };
}
