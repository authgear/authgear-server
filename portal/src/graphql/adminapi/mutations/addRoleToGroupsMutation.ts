import { useMutation } from "@apollo/client";
import { useCallback } from "react";
import {
  AddRoleToGroupsMutationDocument,
  AddRoleToGroupsMutationMutation,
} from "./addRoleToGroupsMutation.generated";

export function useAddRoleToGroupsMutation(): {
  addRoleToGroups: (roleKey: string, groupKeys: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<AddRoleToGroupsMutationMutation>(
      AddRoleToGroupsMutationDocument
    );

  const addRoleToGroups = useCallback(
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

  return { addRoleToGroups, error, loading };
}
