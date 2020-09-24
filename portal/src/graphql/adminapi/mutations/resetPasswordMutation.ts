import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { ResetPasswordMutation } from "./__generated__/ResetPasswordMutation";

const resetPasswordMutation = gql`
  mutation ResetPasswordMutation($userID: ID!, $password: String!) {
    resetPassword(input: { userID: $userID, password: $password }) {
      user {
        id
      }
    }
  }
`;

export function useResetPasswordMutation(
  userID: string
): {
  resetPassword: (password: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] = useMutation<
    ResetPasswordMutation
  >(resetPasswordMutation);

  const resetPassword = useCallback(
    async (password: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
          password,
        },
      });

      return result.data?.resetPassword.user.id ?? null;
    },
    [mutationFunction, userID]
  );

  return { resetPassword, loading, error };
}
