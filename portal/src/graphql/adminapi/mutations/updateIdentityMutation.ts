import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  UpdateIdentityMutationDocument,
  UpdateIdentityMutationMutation,
} from "./updateIdentityMutation.generated";

interface IdentityDefinitionLoginID {
  key: "email" | "phone" | "username";
  value: string;
}

interface Identity {
  id: string;
  claims: Record<string, unknown>;
}

export type UpdateIdentityFunction = (
  identityID: string,
  loginIDIdentity: IdentityDefinitionLoginID
) => Promise<Identity | undefined>;

export function useUpdateLoginIDIdentityMutation(userID: string): {
  updateIdentity: UpdateIdentityFunction;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<UpdateIdentityMutationMutation>(UpdateIdentityMutationDocument);

  const updateIdentity = useCallback(
    async (identityID, loginIDIdentity: IdentityDefinitionLoginID) => {
      const result = await mutationFunction({
        variables: {
          userID,
          identityID,
          definition: { loginID: loginIDIdentity },
        },
      });

      return result.data?.updateIdentity.identity;
    },
    [mutationFunction, userID]
  );
  return { updateIdentity, error, loading };
}
