import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveResourceFromClientIdMutationVariables = Types.Exact<{
  clientID: Types.Scalars['String']['input'];
  resourceURI: Types.Scalars['String']['input'];
}>;


export type RemoveResourceFromClientIdMutation = { __typename?: 'Mutation', removeResourceFromClientID: { __typename?: 'RemoveResourceFromClientIDPayload', resource: { __typename?: 'Resource', id: string, resourceURI: string, name?: string | null } } };


export const RemoveResourceFromClientIdDocument = gql`
    mutation RemoveResourceFromClientID($clientID: String!, $resourceURI: String!) {
  removeResourceFromClientID(
    input: {clientID: $clientID, resourceURI: $resourceURI}
  ) {
    resource {
      id
      resourceURI
      name
    }
  }
}
    `;
export type RemoveResourceFromClientIdMutationFn = Apollo.MutationFunction<RemoveResourceFromClientIdMutation, RemoveResourceFromClientIdMutationVariables>;

/**
 * __useRemoveResourceFromClientIdMutation__
 *
 * To run a mutation, you first call `useRemoveResourceFromClientIdMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveResourceFromClientIdMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeResourceFromClientIdMutation, { data, loading, error }] = useRemoveResourceFromClientIdMutation({
 *   variables: {
 *      clientID: // value for 'clientID'
 *      resourceURI: // value for 'resourceURI'
 *   },
 * });
 */
export function useRemoveResourceFromClientIdMutation(baseOptions?: Apollo.MutationHookOptions<RemoveResourceFromClientIdMutation, RemoveResourceFromClientIdMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveResourceFromClientIdMutation, RemoveResourceFromClientIdMutationVariables>(RemoveResourceFromClientIdDocument, options);
      }
export type RemoveResourceFromClientIdMutationHookResult = ReturnType<typeof useRemoveResourceFromClientIdMutation>;
export type RemoveResourceFromClientIdMutationResult = Apollo.MutationResult<RemoveResourceFromClientIdMutation>;
export type RemoveResourceFromClientIdMutationOptions = Apollo.BaseMutationOptions<RemoveResourceFromClientIdMutation, RemoveResourceFromClientIdMutationVariables>;