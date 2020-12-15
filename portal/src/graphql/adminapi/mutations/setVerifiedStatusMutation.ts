import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { SetVerifiedStatusMutation } from "./__generated__/SetVerifiedStatusMutation";

const setVerifiedStatusMutation = gql`
  mutation SetVerifiedStatusMutation(
    $userID: ID!
    $claimName: String!
    $claimValue: String!
    $isVerified: Boolean!
  ) {
    setVerifiedStatus(
      input: {
        userID: $userID
        claimName: $claimName
        claimValue: $claimValue
        isVerified: $isVerified
      }
    ) {
      user {
        id
        identities {
          edges {
            node {
              id
              claims
            }
          }
        }
        verifiedClaims {
          name
          value
        }
      }
    }
  }
`;

export function useSetVerifiedStatusMutation(
  userID: string
): {
  setVerifiedStatus: (
    claimName: string,
    claimValue: string,
    isVerified: boolean
  ) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [
    mutationFunction,
    { loading, error },
  ] = useMutation<SetVerifiedStatusMutation>(setVerifiedStatusMutation);

  const setVerifiedStatus = useCallback(
    async (claimName: string, claimValue: string, isVerified: boolean) => {
      const result = await mutationFunction({
        variables: {
          userID,
          claimName,
          claimValue,
          isVerified,
        },
      });

      return !!result.data;
    },
    [mutationFunction, userID]
  );

  return { setVerifiedStatus, loading, error };
}
