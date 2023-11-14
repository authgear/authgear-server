import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetVerifiedStatusMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  claimName: Types.Scalars['String']['input'];
  claimValue: Types.Scalars['String']['input'];
  isVerified: Types.Scalars['Boolean']['input'];
}>;


export type SetVerifiedStatusMutationMutation = { __typename?: 'Mutation', setVerifiedStatus: { __typename?: 'SetVerifiedStatusPayload', user: { __typename?: 'User', id: string, identities?: { __typename?: 'IdentityConnection', edges?: Array<{ __typename?: 'IdentityEdge', node?: { __typename?: 'Identity', id: string, claims: any } | null } | null> | null } | null, verifiedClaims: Array<{ __typename?: 'Claim', name: string, value: string }> } } };


export const SetVerifiedStatusMutationDocument = gql`
    mutation setVerifiedStatusMutation($userID: ID!, $claimName: String!, $claimValue: String!, $isVerified: Boolean!) {
  setVerifiedStatus(
    input: {userID: $userID, claimName: $claimName, claimValue: $claimValue, isVerified: $isVerified}
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
export type SetVerifiedStatusMutationMutationFn = Apollo.MutationFunction<SetVerifiedStatusMutationMutation, SetVerifiedStatusMutationMutationVariables>;

/**
 * __useSetVerifiedStatusMutationMutation__
 *
 * To run a mutation, you first call `useSetVerifiedStatusMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetVerifiedStatusMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setVerifiedStatusMutationMutation, { data, loading, error }] = useSetVerifiedStatusMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      claimName: // value for 'claimName'
 *      claimValue: // value for 'claimValue'
 *      isVerified: // value for 'isVerified'
 *   },
 * });
 */
export function useSetVerifiedStatusMutationMutation(baseOptions?: Apollo.MutationHookOptions<SetVerifiedStatusMutationMutation, SetVerifiedStatusMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetVerifiedStatusMutationMutation, SetVerifiedStatusMutationMutationVariables>(SetVerifiedStatusMutationDocument, options);
      }
export type SetVerifiedStatusMutationMutationHookResult = ReturnType<typeof useSetVerifiedStatusMutationMutation>;
export type SetVerifiedStatusMutationMutationResult = Apollo.MutationResult<SetVerifiedStatusMutationMutation>;
export type SetVerifiedStatusMutationMutationOptions = Apollo.BaseMutationOptions<SetVerifiedStatusMutationMutation, SetVerifiedStatusMutationMutationVariables>;