import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { DeleteUserMutation } from "./__generated__/DeleteUserMutation";

const deleteUserMutation = gql`
  mutation DeleteUserMutation($userID: ID!) {
    deleteUser(input: { userID: $userID }) {
      deletedUserID
    }
  }
`;

export function useDeleteUserMutation(): {
  deleteUser: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<DeleteUserMutation>(deleteUserMutation);

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
