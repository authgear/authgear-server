import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  DeleteUserMutationMutation,
  DeleteUserMutationDocument,
} from "./deleteUserMutation.generated";

export function useDeleteUserMutation(): {
  deleteUser: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<DeleteUserMutationMutation>(DeleteUserMutationDocument);

  const deleteUser = useCallback(
    async (userID: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { deleteUser, loading, error };
}
