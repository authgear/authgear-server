import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteScopeMutationMutationVariables = Types.Exact<{
  input: Types.DeleteScopeInput;
}>;


export type DeleteScopeMutationMutation = { __typename?: 'Mutation', deleteScope: { __typename?: 'DeleteScopePayload', ok?: boolean | null } };


export const DeleteScopeMutationDocument = gql`
    mutation DeleteScopeMutation($input: DeleteScopeInput!) {
  deleteScope(input: $input) {
    ok
  }
}
    `;
export type DeleteScopeMutationMutationFn = Apollo.MutationFunction<DeleteScopeMutationMutation, DeleteScopeMutationMutationVariables>;

/**
 * __useDeleteScopeMutationMutation__
 *
 * To run a mutation, you first call `useDeleteScopeMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteScopeMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteScopeMutationMutation, { data, loading, error }] = useDeleteScopeMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useDeleteScopeMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteScopeMutationMutation, DeleteScopeMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteScopeMutationMutation, DeleteScopeMutationMutationVariables>(DeleteScopeMutationDocument, options);
      }
export type DeleteScopeMutationMutationHookResult = ReturnType<typeof useDeleteScopeMutationMutation>;
export type DeleteScopeMutationMutationResult = Apollo.MutationResult<DeleteScopeMutationMutation>;
export type DeleteScopeMutationMutationOptions = Apollo.BaseMutationOptions<DeleteScopeMutationMutation, DeleteScopeMutationMutationVariables>;