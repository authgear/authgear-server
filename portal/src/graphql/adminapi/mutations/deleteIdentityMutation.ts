import { useCallback } from "react";
import { useMutation, gql } from "@apollo/client";
import { DeleteIdentityMutation } from "./__generated__/DeleteIdentityMutation";

const deleteIdentityMutation = gql`
  mutation DeleteIdentityMutation($identityID: ID!) {
    deleteIdentity(input: { identityID: $identityID }) {
      success
    }
  }
`;

export function useDeleteIdentityMutation(): {
  deleteIdentity: (identityID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteIdentityMutation
  >(deleteIdentityMutation, { refetchQueries: ["UserDetailsScreenQuery"] });

  const deleteIdentity = useCallback(
    async (identityID: string) => {
      const result = await mutationFunction({
        variables: {
          identityID,
        },
      });
      return !!result.data?.deleteIdentity.success;
    },
    [mutationFunction]
  );
  return { deleteIdentity, error, loading };
}
