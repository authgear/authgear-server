import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  ResetPasswordMutationMutation,
  ResetPasswordMutationDocument,
} from "./resetPasswordMutation.generated";

export function useResetPasswordMutation(userID: string): {
  resetPassword: (
    password: string,
    sendPassword: boolean,
    forceChangeOnLogin: boolean
  ) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<ResetPasswordMutationMutation>(ResetPasswordMutationDocument);

  const resetPassword = useCallback(
    async (
      password: string,
      sendPassword: boolean,
      forceChangeOnLogin: boolean
    ) => {
      const result = await mutationFunction({
        variables: {
          userID,
          password,
          sendPassword,
          forceChangeOnLogin,
        },
      });

      return result.data?.resetPassword.user.id ?? null;
    },
    [mutationFunction, userID]
  );

  return { resetPassword, loading, error };
}
