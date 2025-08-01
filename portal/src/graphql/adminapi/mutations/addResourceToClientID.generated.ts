import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddResourceToClientIdMutationVariables = Types.Exact<{
  clientID: Types.Scalars['String']['input'];
  resourceURI: Types.Scalars['String']['input'];
}>;


export type AddResourceToClientIdMutation = { __typename?: 'Mutation', addResourceToClientID: { __typename?: 'AddResourceToClientIDPayload', resource: { __typename?: 'Resource', id: string, resourceURI: string, name?: string | null } } };


export const AddResourceToClientIdDocument = gql`
    mutation AddResourceToClientID($clientID: String!, $resourceURI: String!) {
  addResourceToClientID(input: {clientID: $clientID, resourceURI: $resourceURI}) {
    resource {
      id
      resourceURI
      name
    }
  }
}
    `;
export type AddResourceToClientIdMutationFn = Apollo.MutationFunction<AddResourceToClientIdMutation, AddResourceToClientIdMutationVariables>;

/**
 * __useAddResourceToClientIdMutation__
 *
 * To run a mutation, you first call `useAddResourceToClientIdMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddResourceToClientIdMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addResourceToClientIdMutation, { data, loading, error }] = useAddResourceToClientIdMutation({
 *   variables: {
 *      clientID: // value for 'clientID'
 *      resourceURI: // value for 'resourceURI'
 *   },
 * });
 */
export function useAddResourceToClientIdMutation(baseOptions?: Apollo.MutationHookOptions<AddResourceToClientIdMutation, AddResourceToClientIdMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddResourceToClientIdMutation, AddResourceToClientIdMutationVariables>(AddResourceToClientIdDocument, options);
      }
export type AddResourceToClientIdMutationHookResult = ReturnType<typeof useAddResourceToClientIdMutation>;
export type AddResourceToClientIdMutationResult = Apollo.MutationResult<AddResourceToClientIdMutation>;
export type AddResourceToClientIdMutationOptions = Apollo.BaseMutationOptions<AddResourceToClientIdMutation, AddResourceToClientIdMutationVariables>;