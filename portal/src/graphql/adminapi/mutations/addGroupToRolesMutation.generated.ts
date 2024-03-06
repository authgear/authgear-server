import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddGroupToRolesMutationMutationVariables = Types.Exact<{
  groupKey: Types.Scalars['String']['input'];
  roleKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type AddGroupToRolesMutationMutation = { __typename?: 'Mutation', addGroupToRoles: { __typename?: 'AddGroupToRolesPayload', group: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } } };


export const AddGroupToRolesMutationDocument = gql`
    mutation addGroupToRolesMutation($groupKey: String!, $roleKeys: [String!]!) {
  addGroupToRoles(input: {groupKey: $groupKey, roleKeys: $roleKeys}) {
    group {
      id
      key
      name
      description
    }
  }
}
    `;
export type AddGroupToRolesMutationMutationFn = Apollo.MutationFunction<AddGroupToRolesMutationMutation, AddGroupToRolesMutationMutationVariables>;

/**
 * __useAddGroupToRolesMutationMutation__
 *
 * To run a mutation, you first call `useAddGroupToRolesMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddGroupToRolesMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addGroupToRolesMutationMutation, { data, loading, error }] = useAddGroupToRolesMutationMutation({
 *   variables: {
 *      groupKey: // value for 'groupKey'
 *      roleKeys: // value for 'roleKeys'
 *   },
 * });
 */
export function useAddGroupToRolesMutationMutation(baseOptions?: Apollo.MutationHookOptions<AddGroupToRolesMutationMutation, AddGroupToRolesMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddGroupToRolesMutationMutation, AddGroupToRolesMutationMutationVariables>(AddGroupToRolesMutationDocument, options);
      }
export type AddGroupToRolesMutationMutationHookResult = ReturnType<typeof useAddGroupToRolesMutationMutation>;
export type AddGroupToRolesMutationMutationResult = Apollo.MutationResult<AddGroupToRolesMutationMutation>;
export type AddGroupToRolesMutationMutationOptions = Apollo.BaseMutationOptions<AddGroupToRolesMutationMutation, AddGroupToRolesMutationMutationVariables>;