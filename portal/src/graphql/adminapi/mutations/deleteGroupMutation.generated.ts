import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteGroupMutationMutationVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type DeleteGroupMutationMutation = { __typename?: 'Mutation', deleteGroup: { __typename?: 'DeleteGroupPayload', ok?: boolean | null } };


export const DeleteGroupMutationDocument = gql`
    mutation deleteGroupMutation($id: ID!) {
  deleteGroup(input: {id: $id}) {
    ok
  }
}
    `;
export type DeleteGroupMutationMutationFn = Apollo.MutationFunction<DeleteGroupMutationMutation, DeleteGroupMutationMutationVariables>;

/**
 * __useDeleteGroupMutationMutation__
 *
 * To run a mutation, you first call `useDeleteGroupMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteGroupMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteGroupMutationMutation, { data, loading, error }] = useDeleteGroupMutationMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteGroupMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteGroupMutationMutation, DeleteGroupMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteGroupMutationMutation, DeleteGroupMutationMutationVariables>(DeleteGroupMutationDocument, options);
      }
export type DeleteGroupMutationMutationHookResult = ReturnType<typeof useDeleteGroupMutationMutation>;
export type DeleteGroupMutationMutationResult = Apollo.MutationResult<DeleteGroupMutationMutation>;
export type DeleteGroupMutationMutationOptions = Apollo.BaseMutationOptions<DeleteGroupMutationMutation, DeleteGroupMutationMutationVariables>;