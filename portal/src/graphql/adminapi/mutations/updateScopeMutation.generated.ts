import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateScopeMutationMutationVariables = Types.Exact<{
  input: Types.UpdateScopeInput;
}>;


export type UpdateScopeMutationMutation = { __typename?: 'Mutation', updateScope: { __typename?: 'UpdateScopePayload', scope: { __typename?: 'Scope', id: string, scope: string, description?: string | null, resourceID: string, createdAt: any, updatedAt: any } } };


export const UpdateScopeMutationDocument = gql`
    mutation UpdateScopeMutation($input: UpdateScopeInput!) {
  updateScope(input: $input) {
    scope {
      id
      scope
      description
      resourceID
      createdAt
      updatedAt
    }
  }
}
    `;
export type UpdateScopeMutationMutationFn = Apollo.MutationFunction<UpdateScopeMutationMutation, UpdateScopeMutationMutationVariables>;

/**
 * __useUpdateScopeMutationMutation__
 *
 * To run a mutation, you first call `useUpdateScopeMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateScopeMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateScopeMutationMutation, { data, loading, error }] = useUpdateScopeMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateScopeMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateScopeMutationMutation, UpdateScopeMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateScopeMutationMutation, UpdateScopeMutationMutationVariables>(UpdateScopeMutationDocument, options);
      }
export type UpdateScopeMutationMutationHookResult = ReturnType<typeof useUpdateScopeMutationMutation>;
export type UpdateScopeMutationMutationResult = Apollo.MutationResult<UpdateScopeMutationMutation>;
export type UpdateScopeMutationMutationOptions = Apollo.BaseMutationOptions<UpdateScopeMutationMutation, UpdateScopeMutationMutationVariables>;