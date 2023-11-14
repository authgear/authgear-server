import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteAuthorizationMutationMutationVariables = Types.Exact<{
  authorizationID: Types.Scalars['ID']['input'];
}>;


export type DeleteAuthorizationMutationMutation = { __typename?: 'Mutation', deleteAuthorization: { __typename?: 'DeleteAuthorizationPayload', user: { __typename?: 'User', id: string, authorizations?: { __typename?: 'AuthorizationConnection', edges?: Array<{ __typename?: 'AuthorizationEdge', node?: { __typename?: 'Authorization', id: string } | null } | null> | null } | null, sessions?: { __typename?: 'SessionConnection', edges?: Array<{ __typename?: 'SessionEdge', node?: { __typename?: 'Session', id: string } | null } | null> | null } | null } } };


export const DeleteAuthorizationMutationDocument = gql`
    mutation deleteAuthorizationMutation($authorizationID: ID!) {
  deleteAuthorization(input: {authorizationID: $authorizationID}) {
    user {
      id
      authorizations {
        edges {
          node {
            id
          }
        }
      }
      sessions {
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
export type DeleteAuthorizationMutationMutationFn = Apollo.MutationFunction<DeleteAuthorizationMutationMutation, DeleteAuthorizationMutationMutationVariables>;

/**
 * __useDeleteAuthorizationMutationMutation__
 *
 * To run a mutation, you first call `useDeleteAuthorizationMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteAuthorizationMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteAuthorizationMutationMutation, { data, loading, error }] = useDeleteAuthorizationMutationMutation({
 *   variables: {
 *      authorizationID: // value for 'authorizationID'
 *   },
 * });
 */
export function useDeleteAuthorizationMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteAuthorizationMutationMutation, DeleteAuthorizationMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteAuthorizationMutationMutation, DeleteAuthorizationMutationMutationVariables>(DeleteAuthorizationMutationDocument, options);
      }
export type DeleteAuthorizationMutationMutationHookResult = ReturnType<typeof useDeleteAuthorizationMutationMutation>;
export type DeleteAuthorizationMutationMutationResult = Apollo.MutationResult<DeleteAuthorizationMutationMutation>;
export type DeleteAuthorizationMutationMutationOptions = Apollo.BaseMutationOptions<DeleteAuthorizationMutationMutation, DeleteAuthorizationMutationMutationVariables>;