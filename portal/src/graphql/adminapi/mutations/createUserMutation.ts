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
    accountValidFrom?: Date | null;
    accountValidUntil?: Date | null;
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
      accountValidFrom,
      accountValidUntil,
    }: {
      identity: LoginIDIdentity;
      password?: string;
      sendPassword?: boolean;
      setPasswordExpired?: boolean;
      accountValidFrom?: Date | null;
      accountValidUntil?: Date | null;
    }) => {
      const result = await mutationFunction({
        variables: {
          identityDefinition: identity,
          password,
          sendPassword,
          setPasswordExpired,
          accountValidFrom: accountValidFrom ?? null,
          accountValidUntil: accountValidUntil ?? null,
        },
      });
      const userID = result.data?.createUser.user.id ?? null;
      return userID;
    },
    [mutationFunction]
  );

  return { createUser, error, loading };
}
