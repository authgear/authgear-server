import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  CreateIdentityMutationMutation,
  CreateIdentityMutationDocument,
} from "./createIdentityMutation.generated";

interface IdentityDefinitionLoginID {
  key: "email" | "phone" | "username";
  value: string;
}

interface Identity {
  id: string;
  claims: Record<string, unknown>;
}

export type CreateIdentityFunction = (
  loginIDIdentity: IdentityDefinitionLoginID,
  password?: string
) => Promise<Identity | undefined>;

export function useCreateLoginIDIdentityMutation(userID: string): {
  createIdentity: CreateIdentityFunction;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<CreateIdentityMutationMutation>(CreateIdentityMutationDocument);

  const createIdentity = useCallback(
    async (loginIDIdentity: IdentityDefinitionLoginID, password?: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
          definition: { loginID: loginIDIdentity },
          password,
        },
      });

      return result.data?.createIdentity.identity;
    },
    [mutationFunction, userID]
  );
  return { createIdentity, error, loading };
}
