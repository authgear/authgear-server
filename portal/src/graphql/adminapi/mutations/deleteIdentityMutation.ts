import { useCallback } from "react";
import { useMutation, gql } from "@apollo/client";
import { DeleteIdentityMutation } from "./__generated__/DeleteIdentityMutation";

const deleteIdentityMutation = gql`
  mutation DeleteIdentityMutation($identityID: ID!) {
    deleteIdentity(input: { identityID: $identityID }) {
      user {
        id
        authenticators {
          edges {
            node {
              id
            }
          }
        }
        identities {
          edges {
            node {
              id
            }
          }
        }
      }
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
  >(deleteIdentityMutation);

  const deleteIdentity = useCallback(
    async (identityID: string) => {
      const result = await mutationFunction({
        variables: {
          identityID,
        },
      });
      return !!result.data?.deleteIdentity;
    },
    [mutationFunction]
  );
  return { deleteIdentity, error, loading };
}
