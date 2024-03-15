import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveUserFromRolesMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  roleKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type RemoveUserFromRolesMutationMutation = { __typename?: 'Mutation', removeUserFromRoles: { __typename?: 'RemoveUserFromRolesPayload', user: { __typename?: 'User', id: string, formattedName?: string | null } } };


export const RemoveUserFromRolesMutationDocument = gql`
    mutation removeUserFromRolesMutation($userID: ID!, $roleKeys: [String!]!) {
  removeUserFromRoles(input: {userID: $userID, roleKeys: $roleKeys}) {
    user {
      id
      formattedName
    }
  }
}
    `;
export type RemoveUserFromRolesMutationMutationFn = Apollo.MutationFunction<RemoveUserFromRolesMutationMutation, RemoveUserFromRolesMutationMutationVariables>;

/**
 * __useRemoveUserFromRolesMutationMutation__
 *
 * To run a mutation, you first call `useRemoveUserFromRolesMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveUserFromRolesMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeUserFromRolesMutationMutation, { data, loading, error }] = useRemoveUserFromRolesMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      roleKeys: // value for 'roleKeys'
 *   },
 * });
 */
export function useRemoveUserFromRolesMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveUserFromRolesMutationMutation, RemoveUserFromRolesMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveUserFromRolesMutationMutation, RemoveUserFromRolesMutationMutationVariables>(RemoveUserFromRolesMutationDocument, options);
      }
export type RemoveUserFromRolesMutationMutationHookResult = ReturnType<typeof useRemoveUserFromRolesMutationMutation>;
export type RemoveUserFromRolesMutationMutationResult = Apollo.MutationResult<RemoveUserFromRolesMutationMutation>;
export type RemoveUserFromRolesMutationMutationOptions = Apollo.BaseMutationOptions<RemoveUserFromRolesMutationMutation, RemoveUserFromRolesMutationMutationVariables>;