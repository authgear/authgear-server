import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateScopeMutationMutationVariables = Types.Exact<{
  input: Types.CreateScopeInput;
}>;


export type CreateScopeMutationMutation = { __typename?: 'Mutation', createScope: { __typename?: 'CreateScopePayload', scope: { __typename?: 'Scope', id: string, scope: string, description?: string | null, resourceID: string, createdAt: any, updatedAt: any } } };


export const CreateScopeMutationDocument = gql`
    mutation CreateScopeMutation($input: CreateScopeInput!) {
  createScope(input: $input) {
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
export type CreateScopeMutationMutationFn = Apollo.MutationFunction<CreateScopeMutationMutation, CreateScopeMutationMutationVariables>;

/**
 * __useCreateScopeMutationMutation__
 *
 * To run a mutation, you first call `useCreateScopeMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateScopeMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createScopeMutationMutation, { data, loading, error }] = useCreateScopeMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateScopeMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateScopeMutationMutation, CreateScopeMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateScopeMutationMutation, CreateScopeMutationMutationVariables>(CreateScopeMutationDocument, options);
      }
export type CreateScopeMutationMutationHookResult = ReturnType<typeof useCreateScopeMutationMutation>;
export type CreateScopeMutationMutationResult = Apollo.MutationResult<CreateScopeMutationMutation>;
export type CreateScopeMutationMutationOptions = Apollo.BaseMutationOptions<CreateScopeMutationMutation, CreateScopeMutationMutationVariables>;