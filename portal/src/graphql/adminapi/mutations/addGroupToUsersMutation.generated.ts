import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddGroupToUsersMutationMutationVariables = Types.Exact<{
  groupKey: Types.Scalars['String']['input'];
  userIDs: Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input'];
}>;


export type AddGroupToUsersMutationMutation = { __typename?: 'Mutation', addGroupToUsers: { __typename?: 'AddGroupToUsersPayload', group: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } } };


export const AddGroupToUsersMutationDocument = gql`
    mutation addGroupToUsersMutation($groupKey: String!, $userIDs: [ID!]!) {
  addGroupToUsers(input: {groupKey: $groupKey, userIDs: $userIDs}) {
    group {
      id
      key
      name
      description
    }
  }
}
    `;
export type AddGroupToUsersMutationMutationFn = Apollo.MutationFunction<AddGroupToUsersMutationMutation, AddGroupToUsersMutationMutationVariables>;

/**
 * __useAddGroupToUsersMutationMutation__
 *
 * To run a mutation, you first call `useAddGroupToUsersMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddGroupToUsersMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addGroupToUsersMutationMutation, { data, loading, error }] = useAddGroupToUsersMutationMutation({
 *   variables: {
 *      groupKey: // value for 'groupKey'
 *      userIDs: // value for 'userIDs'
 *   },
 * });
 */
export function useAddGroupToUsersMutationMutation(baseOptions?: Apollo.MutationHookOptions<AddGroupToUsersMutationMutation, AddGroupToUsersMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddGroupToUsersMutationMutation, AddGroupToUsersMutationMutationVariables>(AddGroupToUsersMutationDocument, options);
      }
export type AddGroupToUsersMutationMutationHookResult = ReturnType<typeof useAddGroupToUsersMutationMutation>;
export type AddGroupToUsersMutationMutationResult = Apollo.MutationResult<AddGroupToUsersMutationMutation>;
export type AddGroupToUsersMutationMutationOptions = Apollo.BaseMutationOptions<AddGroupToUsersMutationMutation, AddGroupToUsersMutationMutationVariables>;