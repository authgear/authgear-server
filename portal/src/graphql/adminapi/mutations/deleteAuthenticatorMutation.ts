import { gql, useMutation } from "@apollo/client";
import { useCallback } from "react";
import { DeleteAuthenticatorMutation } from "./__generated__/DeleteAuthenticatorMutation";

const deleteAuthenticatorMutation = gql`
  mutation DeleteAuthenticatorMutation($authenticatorID: ID!) {
    deleteAuthenticator(input: { authenticatorID: $authenticatorID }) {
      success
    }
  }
`;

export function useDeleteAuthenticatorMutation(): {
  deleteAuthenticator: (authenticatorID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteAuthenticatorMutation
  >(deleteAuthenticatorMutation);
  const deleteAuthenticator = useCallback(
    async (authenticatorID: string) => {
      const result = await mutationFunction({
        variables: {
          authenticatorID,
        },
        refetchQueries: ["UserQuery"],
      });

      return result.data?.deleteAuthenticator.success ?? false;
    },
    [mutationFunction]
  );

  return {
    deleteAuthenticator,
    loading,
    error,
  };
}
