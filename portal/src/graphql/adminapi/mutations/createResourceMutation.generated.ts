import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateResourceMutationMutationVariables = Types.Exact<{
  input: Types.CreateResourceInput;
}>;


export type CreateResourceMutationMutation = { __typename?: 'Mutation', createResource: { __typename?: 'CreateResourcePayload', resource: { __typename?: 'Resource', id: string, name?: string | null, resourceURI: string, createdAt: any, updatedAt: any } } };


export const CreateResourceMutationDocument = gql`
    mutation CreateResourceMutation($input: CreateResourceInput!) {
  createResource(input: $input) {
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
export type CreateResourceMutationMutationFn = Apollo.MutationFunction<CreateResourceMutationMutation, CreateResourceMutationMutationVariables>;

/**
 * __useCreateResourceMutationMutation__
 *
 * To run a mutation, you first call `useCreateResourceMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateResourceMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createResourceMutationMutation, { data, loading, error }] = useCreateResourceMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateResourceMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateResourceMutationMutation, CreateResourceMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateResourceMutationMutation, CreateResourceMutationMutationVariables>(CreateResourceMutationDocument, options);
      }
export type CreateResourceMutationMutationHookResult = ReturnType<typeof useCreateResourceMutationMutation>;
export type CreateResourceMutationMutationResult = Apollo.MutationResult<CreateResourceMutationMutation>;
export type CreateResourceMutationMutationOptions = Apollo.BaseMutationOptions<CreateResourceMutationMutation, CreateResourceMutationMutationVariables>;