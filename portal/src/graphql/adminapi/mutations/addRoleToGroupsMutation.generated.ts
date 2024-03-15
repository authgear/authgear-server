import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddRoleToGroupsMutationMutationVariables = Types.Exact<{
  roleKey: Types.Scalars['String']['input'];
  groupKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type AddRoleToGroupsMutationMutation = { __typename?: 'Mutation', addRoleToGroups: { __typename?: 'AddRoleToGroupsPayload', role: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } } };


export const AddRoleToGroupsMutationDocument = gql`
    mutation addRoleToGroupsMutation($roleKey: String!, $groupKeys: [String!]!) {
  addRoleToGroups(input: {roleKey: $roleKey, groupKeys: $groupKeys}) {
    role {
      id
      key
      name
      description
    }
  }
}
    `;
export type AddRoleToGroupsMutationMutationFn = Apollo.MutationFunction<AddRoleToGroupsMutationMutation, AddRoleToGroupsMutationMutationVariables>;

/**
 * __useAddRoleToGroupsMutationMutation__
 *
 * To run a mutation, you first call `useAddRoleToGroupsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddRoleToGroupsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addRoleToGroupsMutationMutation, { data, loading, error }] = useAddRoleToGroupsMutationMutation({
 *   variables: {
 *      roleKey: // value for 'roleKey'
 *      groupKeys: // value for 'groupKeys'
 *   },
 * });
 */
export function useAddRoleToGroupsMutationMutation(baseOptions?: Apollo.MutationHookOptions<AddRoleToGroupsMutationMutation, AddRoleToGroupsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddRoleToGroupsMutationMutation, AddRoleToGroupsMutationMutationVariables>(AddRoleToGroupsMutationDocument, options);
      }
export type AddRoleToGroupsMutationMutationHookResult = ReturnType<typeof useAddRoleToGroupsMutationMutation>;
export type AddRoleToGroupsMutationMutationResult = Apollo.MutationResult<AddRoleToGroupsMutationMutation>;
export type AddRoleToGroupsMutationMutationOptions = Apollo.BaseMutationOptions<AddRoleToGroupsMutationMutation, AddRoleToGroupsMutationMutationVariables>;