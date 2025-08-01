import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateResourceMutationMutationVariables = Types.Exact<{
  input: Types.UpdateResourceInput;
}>;


export type UpdateResourceMutationMutation = { __typename?: 'Mutation', updateResource: { __typename?: 'UpdateResourcePayload', resource: { __typename?: 'Resource', id: string, name?: string | null, resourceURI: string, createdAt: any, updatedAt: any } } };


export const UpdateResourceMutationDocument = gql`
    mutation UpdateResourceMutation($input: UpdateResourceInput!) {
  updateResource(input: $input) {
    resource {
      id
      name
      resourceURI
      createdAt
      updatedAt
    }
  }
}
    `;
export type UpdateResourceMutationMutationFn = Apollo.MutationFunction<UpdateResourceMutationMutation, UpdateResourceMutationMutationVariables>;

/**
 * __useUpdateResourceMutationMutation__
 *
 * To run a mutation, you first call `useUpdateResourceMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateResourceMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateResourceMutationMutation, { data, loading, error }] = useUpdateResourceMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateResourceMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateResourceMutationMutation, UpdateResourceMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateResourceMutationMutation, UpdateResourceMutationMutationVariables>(UpdateResourceMutationDocument, options);
      }
export type UpdateResourceMutationMutationHookResult = ReturnType<typeof useUpdateResourceMutationMutation>;
export type UpdateResourceMutationMutationResult = Apollo.MutationResult<UpdateResourceMutationMutation>;
export type UpdateResourceMutationMutationOptions = Apollo.BaseMutationOptions<UpdateResourceMutationMutation, UpdateResourceMutationMutationVariables>;