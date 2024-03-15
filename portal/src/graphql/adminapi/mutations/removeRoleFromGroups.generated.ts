import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveRoleFromGroupsMutationMutationVariables = Types.Exact<{
  roleKey: Types.Scalars['String']['input'];
  groupKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type RemoveRoleFromGroupsMutationMutation = { __typename?: 'Mutation', removeRoleFromGroups: { __typename?: 'RemoveRoleFromGroupsPayload', role: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } } };


export const RemoveRoleFromGroupsMutationDocument = gql`
    mutation removeRoleFromGroupsMutation($roleKey: String!, $groupKeys: [String!]!) {
  removeRoleFromGroups(input: {roleKey: $roleKey, groupKeys: $groupKeys}) {
    role {
      id
      key
      name
      description
    }
  }
}
    `;
export type RemoveRoleFromGroupsMutationMutationFn = Apollo.MutationFunction<RemoveRoleFromGroupsMutationMutation, RemoveRoleFromGroupsMutationMutationVariables>;

/**
 * __useRemoveRoleFromGroupsMutationMutation__
 *
 * To run a mutation, you first call `useRemoveRoleFromGroupsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveRoleFromGroupsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeRoleFromGroupsMutationMutation, { data, loading, error }] = useRemoveRoleFromGroupsMutationMutation({
 *   variables: {
 *      roleKey: // value for 'roleKey'
 *      groupKeys: // value for 'groupKeys'
 *   },
 * });
 */
export function useRemoveRoleFromGroupsMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveRoleFromGroupsMutationMutation, RemoveRoleFromGroupsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveRoleFromGroupsMutationMutation, RemoveRoleFromGroupsMutationMutationVariables>(RemoveRoleFromGroupsMutationDocument, options);
      }
export type RemoveRoleFromGroupsMutationMutationHookResult = ReturnType<typeof useRemoveRoleFromGroupsMutationMutation>;
export type RemoveRoleFromGroupsMutationMutationResult = Apollo.MutationResult<RemoveRoleFromGroupsMutationMutation>;
export type RemoveRoleFromGroupsMutationMutationOptions = Apollo.BaseMutationOptions<RemoveRoleFromGroupsMutationMutation, RemoveRoleFromGroupsMutationMutationVariables>;