import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteDynamicClientMutationMutationVariables = Types.Exact<{
  input: Types.DeleteDynamicClientInput;
}>;


export type DeleteDynamicClientMutationMutation = { __typename?: 'Mutation', deleteDynamicClient: { __typename?: 'DeleteDynamicClientPayload', ok?: boolean | null } };


export const DeleteDynamicClientMutationDocument = gql`
    mutation DeleteDynamicClientMutation($input: DeleteDynamicClientInput!) {
  deleteDynamicClient(input: $input) {
    ok
  }
}
    `;
export type DeleteDynamicClientMutationMutationFn = Apollo.MutationFunction<DeleteDynamicClientMutationMutation, DeleteDynamicClientMutationMutationVariables>;

/**
 * __useDeleteDynamicClientMutationMutation__
 *
 * To run a mutation, you first call `useDeleteDynamicClientMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteDynamicClientMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteDynamicClientMutationMutation, { data, loading, error }] = useDeleteDynamicClientMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useDeleteDynamicClientMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteDynamicClientMutationMutation, DeleteDynamicClientMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteDynamicClientMutationMutation, DeleteDynamicClientMutationMutationVariables>(DeleteDynamicClientMutationDocument, options);
      }
export type DeleteDynamicClientMutationMutationHookResult = ReturnType<typeof useDeleteDynamicClientMutationMutation>;
export type DeleteDynamicClientMutationMutationResult = Apollo.MutationResult<DeleteDynamicClientMutationMutation>;
export type DeleteDynamicClientMutationMutationOptions = Apollo.BaseMutationOptions<DeleteDynamicClientMutationMutation, DeleteDynamicClientMutationMutationVariables>;