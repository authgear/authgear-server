import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteIdentityMutationMutationVariables = Types.Exact<{
  identityID: Types.Scalars['ID']['input'];
}>;


export type DeleteIdentityMutationMutation = { __typename?: 'Mutation', deleteIdentity: { __typename?: 'DeleteIdentityPayload', user: { __typename?: 'User', id: string, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string } | null } | null> | null } | null, identities?: { __typename?: 'IdentityConnection', edges?: Array<{ __typename?: 'IdentityEdge', node?: { __typename?: 'Identity', id: string } | null } | null> | null } | null } } };


export const DeleteIdentityMutationDocument = gql`
    mutation deleteIdentityMutation($identityID: ID!) {
  deleteIdentity(input: {identityID: $identityID}) {
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
export type DeleteIdentityMutationMutationFn = Apollo.MutationFunction<DeleteIdentityMutationMutation, DeleteIdentityMutationMutationVariables>;

/**
 * __useDeleteIdentityMutationMutation__
 *
 * To run a mutation, you first call `useDeleteIdentityMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteIdentityMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteIdentityMutationMutation, { data, loading, error }] = useDeleteIdentityMutationMutation({
 *   variables: {
 *      identityID: // value for 'identityID'
 *   },
 * });
 */
export function useDeleteIdentityMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteIdentityMutationMutation, DeleteIdentityMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteIdentityMutationMutation, DeleteIdentityMutationMutationVariables>(DeleteIdentityMutationDocument, options);
      }
export type DeleteIdentityMutationMutationHookResult = ReturnType<typeof useDeleteIdentityMutationMutation>;
export type DeleteIdentityMutationMutationResult = Apollo.MutationResult<DeleteIdentityMutationMutation>;
export type DeleteIdentityMutationMutationOptions = Apollo.BaseMutationOptions<DeleteIdentityMutationMutation, DeleteIdentityMutationMutationVariables>;