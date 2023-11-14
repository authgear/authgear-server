import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GenerateTesterTokenMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  returnUri: Types.Scalars['String']['input'];
}>;


export type GenerateTesterTokenMutationMutation = { __typename?: 'Mutation', generateTesterToken: { __typename?: 'GenerateTestTokenPayload', token: string } };


export const GenerateTesterTokenMutationDocument = gql`
    mutation generateTesterTokenMutation($appID: ID!, $returnUri: String!) {
  generateTesterToken(input: {id: $appID, returnUri: $returnUri}) {
    token
  }
}
    `;
export type GenerateTesterTokenMutationMutationFn = Apollo.MutationFunction<GenerateTesterTokenMutationMutation, GenerateTesterTokenMutationMutationVariables>;

/**
 * __useGenerateTesterTokenMutationMutation__
 *
 * To run a mutation, you first call `useGenerateTesterTokenMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useGenerateTesterTokenMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [generateTesterTokenMutationMutation, { data, loading, error }] = useGenerateTesterTokenMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      returnUri: // value for 'returnUri'
 *   },
 * });
 */
export function useGenerateTesterTokenMutationMutation(baseOptions?: Apollo.MutationHookOptions<GenerateTesterTokenMutationMutation, GenerateTesterTokenMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<GenerateTesterTokenMutationMutation, GenerateTesterTokenMutationMutationVariables>(GenerateTesterTokenMutationDocument, options);
      }
export type GenerateTesterTokenMutationMutationHookResult = ReturnType<typeof useGenerateTesterTokenMutationMutation>;
export type GenerateTesterTokenMutationMutationResult = Apollo.MutationResult<GenerateTesterTokenMutationMutation>;
export type GenerateTesterTokenMutationMutationOptions = Apollo.BaseMutationOptions<GenerateTesterTokenMutationMutation, GenerateTesterTokenMutationMutationVariables>;