import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteRoleMutationMutationVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type DeleteRoleMutationMutation = { __typename?: 'Mutation', deleteRole: { __typename?: 'DeleteRolePayload', ok?: boolean | null } };


export const DeleteRoleMutationDocument = gql`
    mutation deleteRoleMutation($id: ID!) {
  deleteRole(input: {id: $id}) {
    ok
  }
}
    `;
export type DeleteRoleMutationMutationFn = Apollo.MutationFunction<DeleteRoleMutationMutation, DeleteRoleMutationMutationVariables>;

/**
 * __useDeleteRoleMutationMutation__
 *
 * To run a mutation, you first call `useDeleteRoleMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteRoleMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteRoleMutationMutation, { data, loading, error }] = useDeleteRoleMutationMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteRoleMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteRoleMutationMutation, DeleteRoleMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteRoleMutationMutation, DeleteRoleMutationMutationVariables>(DeleteRoleMutationDocument, options);
      }
export type DeleteRoleMutationMutationHookResult = ReturnType<typeof useDeleteRoleMutationMutation>;
export type DeleteRoleMutationMutationResult = Apollo.MutationResult<DeleteRoleMutationMutation>;
export type DeleteRoleMutationMutationOptions = Apollo.BaseMutationOptions<DeleteRoleMutationMutation, DeleteRoleMutationMutationVariables>;