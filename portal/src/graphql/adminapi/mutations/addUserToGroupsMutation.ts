import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  AddUserToGroupsMutationDocument,
  AddUserToGroupsMutationMutation,
} from "./addUserToGroupsMutation.generated";

export function useAddUserToGroupsMutation(): {
  addUserToGroups: (userID: string, groupKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<AddUserToGroupsMutationMutation>(
      AddUserToGroupsMutationDocument
    );

  const addUserToGroups = useCallback(
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

  return { addUserToGroups, error, loading };
}
