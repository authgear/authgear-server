import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CheckDenoHookMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  content: Types.Scalars['String']['input'];
}>;


export type CheckDenoHookMutationMutation = { __typename?: 'Mutation', checkDenoHook?: boolean | null };


export const CheckDenoHookMutationDocument = gql`
    mutation checkDenoHookMutation($appID: ID!, $content: String!) {
  checkDenoHook(input: {appID: $appID, content: $content})
}
    `;
export type CheckDenoHookMutationMutationFn = Apollo.MutationFunction<CheckDenoHookMutationMutation, CheckDenoHookMutationMutationVariables>;

/**
 * __useCheckDenoHookMutationMutation__
 *
 * To run a mutation, you first call `useCheckDenoHookMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCheckDenoHookMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [checkDenoHookMutationMutation, { data, loading, error }] = useCheckDenoHookMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      content: // value for 'content'
 *   },
 * });
 */
export function useCheckDenoHookMutationMutation(baseOptions?: Apollo.MutationHookOptions<CheckDenoHookMutationMutation, CheckDenoHookMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CheckDenoHookMutationMutation, CheckDenoHookMutationMutationVariables>(CheckDenoHookMutationDocument, options);
      }
export type CheckDenoHookMutationMutationHookResult = ReturnType<typeof useCheckDenoHookMutationMutation>;
export type CheckDenoHookMutationMutationResult = Apollo.MutationResult<CheckDenoHookMutationMutation>;
export type CheckDenoHookMutationMutationOptions = Apollo.BaseMutationOptions<CheckDenoHookMutationMutation, CheckDenoHookMutationMutationVariables>;