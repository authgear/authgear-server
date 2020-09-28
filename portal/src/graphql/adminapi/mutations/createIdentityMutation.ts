import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { CreateIdentityMutation } from "./__generated__/CreateIdentityMutation";

const createIdentityMutation = gql`
  mutation CreateIdentityMutation(
    $userID: ID!
    $definition: IdentityDefinition!
    $password: String
  ) {
    createIdentity(
      input: { userID: $userID, definition: $definition, password: $password }
    ) {
      user {
        id
        identities {
          edges {
            node {
              id
            }
          }
        }
      }
      identity {
        id
        type
        claims
        createdAt
        updatedAt
      }
    }
  }
`;

interface IdentityDefinitionLoginID {
  key: "email" | "phone" | "username";
  value: string;
}

interface Identity {
  id: string;
  claims: Record<string, unknown>;
}

export function useCreateLoginIDIdentityMutation(
  userID: string
): {
  createIdentity: (
    loginIDIdentity: IdentityDefinitionLoginID,
    password?: string
  ) => Promise<Identity | undefined>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    CreateIdentityMutation
  >(createIdentityMutation);

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
