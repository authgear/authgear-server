import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveGroupFromUsersMutationMutationVariables = Types.Exact<{
  groupKey: Types.Scalars['String']['input'];
  userIDs: Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input'];
}>;


export type RemoveGroupFromUsersMutationMutation = { __typename?: 'Mutation', removeGroupFromUsers: { __typename?: 'RemoveGroupToUsersPayload', group: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } } };


export const RemoveGroupFromUsersMutationDocument = gql`
    mutation removeGroupFromUsersMutation($groupKey: String!, $userIDs: [ID!]!) {
  removeGroupFromUsers(input: {groupKey: $groupKey, userIDs: $userIDs}) {
    group {
      id
      key
      name
      description
    }
  }
}
    `;
export type RemoveGroupFromUsersMutationMutationFn = Apollo.MutationFunction<RemoveGroupFromUsersMutationMutation, RemoveGroupFromUsersMutationMutationVariables>;

/**
 * __useRemoveGroupFromUsersMutationMutation__
 *
 * To run a mutation, you first call `useRemoveGroupFromUsersMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveGroupFromUsersMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeGroupFromUsersMutationMutation, { data, loading, error }] = useRemoveGroupFromUsersMutationMutation({
 *   variables: {
 *      groupKey: // value for 'groupKey'
 *      userIDs: // value for 'userIDs'
 *   },
 * });
 */
export function useRemoveGroupFromUsersMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveGroupFromUsersMutationMutation, RemoveGroupFromUsersMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveGroupFromUsersMutationMutation, RemoveGroupFromUsersMutationMutationVariables>(RemoveGroupFromUsersMutationDocument, options);
      }
export type RemoveGroupFromUsersMutationMutationHookResult = ReturnType<typeof useRemoveGroupFromUsersMutationMutation>;
export type RemoveGroupFromUsersMutationMutationResult = Apollo.MutationResult<RemoveGroupFromUsersMutationMutation>;
export type RemoveGroupFromUsersMutationMutationOptions = Apollo.BaseMutationOptions<RemoveGroupFromUsersMutationMutation, RemoveGroupFromUsersMutationMutationVariables>;