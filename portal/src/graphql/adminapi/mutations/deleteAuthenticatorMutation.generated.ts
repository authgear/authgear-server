import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteAuthenticatorMutationMutationVariables = Types.Exact<{
  authenticatorID: Types.Scalars['ID']['input'];
}>;


export type DeleteAuthenticatorMutationMutation = { __typename?: 'Mutation', deleteAuthenticator: { __typename?: 'DeleteAuthenticatorPayload', user: { __typename?: 'User', id: string, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string } | null } | null> | null } | null } } };


export const DeleteAuthenticatorMutationDocument = gql`
    mutation deleteAuthenticatorMutation($authenticatorID: ID!) {
  deleteAuthenticator(input: {authenticatorID: $authenticatorID}) {
    user {
      id
      authenticators {
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
export type DeleteAuthenticatorMutationMutationFn = Apollo.MutationFunction<DeleteAuthenticatorMutationMutation, DeleteAuthenticatorMutationMutationVariables>;

/**
 * __useDeleteAuthenticatorMutationMutation__
 *
 * To run a mutation, you first call `useDeleteAuthenticatorMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteAuthenticatorMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteAuthenticatorMutationMutation, { data, loading, error }] = useDeleteAuthenticatorMutationMutation({
 *   variables: {
 *      authenticatorID: // value for 'authenticatorID'
 *   },
 * });
 */
export function useDeleteAuthenticatorMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteAuthenticatorMutationMutation, DeleteAuthenticatorMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteAuthenticatorMutationMutation, DeleteAuthenticatorMutationMutationVariables>(DeleteAuthenticatorMutationDocument, options);
      }
export type DeleteAuthenticatorMutationMutationHookResult = ReturnType<typeof useDeleteAuthenticatorMutationMutation>;
export type DeleteAuthenticatorMutationMutationResult = Apollo.MutationResult<DeleteAuthenticatorMutationMutation>;
export type DeleteAuthenticatorMutationMutationOptions = Apollo.BaseMutationOptions<DeleteAuthenticatorMutationMutation, DeleteAuthenticatorMutationMutationVariables>;