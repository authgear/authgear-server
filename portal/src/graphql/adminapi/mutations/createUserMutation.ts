import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { CreateUserMutation } from "./__generated__/CreateUserMutation";

const createUserMutation = gql`
  mutation CreateUserMutation(
    $identityDefinition: IdentityDefinitionLoginID!
    $password: String
  ) {
    createUser(
      input: {
        definition: { loginID: $identityDefinition }
        password: $password
      }
    ) {
      user {
        id
      }
    }
  }
`;

interface Identity {
  key: "username" | "email" | "phone";
  value: string;
}

export function useCreateUserMutation(): {
  createUser: (identity: Identity, password?: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    CreateUserMutation
  >(createUserMutation);
  const createUser = useCallback(
    async (identity: Identity, password?: string) => {
      const result = await mutationFunction({
        variables: {
          identityDefinition: identity,
          password,
        },
      });
      const userID = result.data?.createUser.user.id ?? null;
      return userID;
    },
    [mutationFunction]
  );

  return { createUser, error, loading };
}
