import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteUserMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type DeleteUserMutationMutation = { __typename?: 'Mutation', deleteUser: { __typename?: 'DeleteUserPayload', deletedUserID: string } };


export const DeleteUserMutationDocument = gql`
    mutation deleteUserMutation($userID: ID!) {
  deleteUser(input: {userID: $userID}) {
    deletedUserID
  }
}
    `;
export type DeleteUserMutationMutationFn = Apollo.MutationFunction<DeleteUserMutationMutation, DeleteUserMutationMutationVariables>;

/**
 * __useDeleteUserMutationMutation__
 *
 * To run a mutation, you first call `useDeleteUserMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteUserMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteUserMutationMutation, { data, loading, error }] = useDeleteUserMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useDeleteUserMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteUserMutationMutation, DeleteUserMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteUserMutationMutation, DeleteUserMutationMutationVariables>(DeleteUserMutationDocument, options);
      }
export type DeleteUserMutationMutationHookResult = ReturnType<typeof useDeleteUserMutationMutation>;
export type DeleteUserMutationMutationResult = Apollo.MutationResult<DeleteUserMutationMutation>;
export type DeleteUserMutationMutationOptions = Apollo.BaseMutationOptions<DeleteUserMutationMutation, DeleteUserMutationMutationVariables>;