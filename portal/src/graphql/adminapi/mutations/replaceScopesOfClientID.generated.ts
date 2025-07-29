import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ReplaceScopesOfClientIdMutationVariables = Types.Exact<{
  clientID: Types.Scalars['String']['input'];
  resourceURI: Types.Scalars['String']['input'];
  scopes: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type ReplaceScopesOfClientIdMutation = { __typename?: 'Mutation', replaceScopesOfClientID: { __typename?: 'ReplaceScopesOfClientIDPayload', scopes: Array<{ __typename?: 'Scope', id: string, scope: string, resourceID: string, description?: string | null, createdAt: any, updatedAt: any }> } };


export const ReplaceScopesOfClientIdDocument = gql`
    mutation ReplaceScopesOfClientID($clientID: String!, $resourceURI: String!, $scopes: [String!]!) {
  replaceScopesOfClientID(
    input: {clientID: $clientID, resourceURI: $resourceURI, scopes: $scopes}
  ) {
    scopes {
      id
      scope
      resourceID
      description
      createdAt
      updatedAt
    }
  }
}
    `;
export type ReplaceScopesOfClientIdMutationFn = Apollo.MutationFunction<ReplaceScopesOfClientIdMutation, ReplaceScopesOfClientIdMutationVariables>;

/**
 * __useReplaceScopesOfClientIdMutation__
 *
 * To run a mutation, you first call `useReplaceScopesOfClientIdMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useReplaceScopesOfClientIdMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [replaceScopesOfClientIdMutation, { data, loading, error }] = useReplaceScopesOfClientIdMutation({
 *   variables: {
 *      clientID: // value for 'clientID'
 *      resourceURI: // value for 'resourceURI'
 *      scopes: // value for 'scopes'
 *   },
 * });
 */
export function useReplaceScopesOfClientIdMutation(baseOptions?: Apollo.MutationHookOptions<ReplaceScopesOfClientIdMutation, ReplaceScopesOfClientIdMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ReplaceScopesOfClientIdMutation, ReplaceScopesOfClientIdMutationVariables>(ReplaceScopesOfClientIdDocument, options);
      }
export type ReplaceScopesOfClientIdMutationHookResult = ReturnType<typeof useReplaceScopesOfClientIdMutation>;
export type ReplaceScopesOfClientIdMutationResult = Apollo.MutationResult<ReplaceScopesOfClientIdMutation>;
export type ReplaceScopesOfClientIdMutationOptions = Apollo.BaseMutationOptions<ReplaceScopesOfClientIdMutation, ReplaceScopesOfClientIdMutationVariables>;