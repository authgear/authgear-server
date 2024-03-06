import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveGroupFromRolesMutationMutationVariables = Types.Exact<{
  groupKey: Types.Scalars['String']['input'];
  roleKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type RemoveGroupFromRolesMutationMutation = { __typename?: 'Mutation', removeGroupFromRoles: { __typename?: 'RemoveGroupFromRolesPayload', group: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } } };


export const RemoveGroupFromRolesMutationDocument = gql`
    mutation removeGroupFromRolesMutation($groupKey: String!, $roleKeys: [String!]!) {
  removeGroupFromRoles(input: {groupKey: $groupKey, roleKeys: $roleKeys}) {
    group {
      id
      key
      name
      description
    }
  }
}
    `;
export type RemoveGroupFromRolesMutationMutationFn = Apollo.MutationFunction<RemoveGroupFromRolesMutationMutation, RemoveGroupFromRolesMutationMutationVariables>;

/**
 * __useRemoveGroupFromRolesMutationMutation__
 *
 * To run a mutation, you first call `useRemoveGroupFromRolesMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveGroupFromRolesMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeGroupFromRolesMutationMutation, { data, loading, error }] = useRemoveGroupFromRolesMutationMutation({
 *   variables: {
 *      groupKey: // value for 'groupKey'
 *      roleKeys: // value for 'roleKeys'
 *   },
 * });
 */
export function useRemoveGroupFromRolesMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveGroupFromRolesMutationMutation, RemoveGroupFromRolesMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveGroupFromRolesMutationMutation, RemoveGroupFromRolesMutationMutationVariables>(RemoveGroupFromRolesMutationDocument, options);
      }
export type RemoveGroupFromRolesMutationMutationHookResult = ReturnType<typeof useRemoveGroupFromRolesMutationMutation>;
export type RemoveGroupFromRolesMutationMutationResult = Apollo.MutationResult<RemoveGroupFromRolesMutationMutation>;
export type RemoveGroupFromRolesMutationMutationOptions = Apollo.BaseMutationOptions<RemoveGroupFromRolesMutationMutation, RemoveGroupFromRolesMutationMutationVariables>;