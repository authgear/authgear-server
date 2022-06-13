import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  ResetPasswordMutationMutation,
  ResetPasswordMutationDocument,
} from "./resetPasswordMutation.generated";

export function useResetPasswordMutation(userID: string): {
  resetPassword: (password: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<ResetPasswordMutationMutation>(ResetPasswordMutationDocument);

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
