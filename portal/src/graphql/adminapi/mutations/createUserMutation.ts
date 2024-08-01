import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  CreateUserMutationMutation,
  CreateUserMutationDocument,
} from "./createUserMutation.generated";

interface LoginIDIdentity {
  key: "username" | "email" | "phone";
  value: string;
}

export function useCreateUserMutation(): {
  createUser: (input: {
    identity: LoginIDIdentity;
    password?: string;
    sendPassword?: boolean;
    setPasswordExpired?: boolean;
  }) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<CreateUserMutationMutation>(CreateUserMutationDocument);
  const createUser = useCallback(
    async ({
      identity,
      password,
      sendPassword,
      setPasswordExpired,
    }: {
      identity: LoginIDIdentity;
      password?: string;
      sendPassword?: boolean;
      setPasswordExpired?: boolean;
    }) => {
      const result = await mutationFunction({
        variables: {
          identityDefinition: identity,
          password,
          sendPassword,
          setPasswordExpired,
        },
      });
      const userID = result.data?.createUser.user.id ?? null;
      return userID;
    },
    [mutationFunction]
  );

  return { createUser, error, loading };
}
