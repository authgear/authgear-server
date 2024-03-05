import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddUserToRolesMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  roleKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type AddUserToRolesMutationMutation = { __typename?: 'Mutation', addUserToRoles: { __typename?: 'AddUserToRolesPayload', user: { __typename?: 'User', id: string, formattedName?: string | null } } };


export const AddUserToRolesMutationDocument = gql`
    mutation addUserToRolesMutation($userID: ID!, $roleKeys: [String!]!) {
  addUserToRoles(input: {userID: $userID, roleKeys: $roleKeys}) {
    user {
      id
      formattedName
    }
  }
}
    `;
export type AddUserToRolesMutationMutationFn = Apollo.MutationFunction<AddUserToRolesMutationMutation, AddUserToRolesMutationMutationVariables>;

/**
 * __useAddUserToRolesMutationMutation__
 *
 * To run a mutation, you first call `useAddUserToRolesMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddUserToRolesMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addUserToRolesMutationMutation, { data, loading, error }] = useAddUserToRolesMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      roleKeys: // value for 'roleKeys'
 *   },
 * });
 */
export function useAddUserToRolesMutationMutation(baseOptions?: Apollo.MutationHookOptions<AddUserToRolesMutationMutation, AddUserToRolesMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddUserToRolesMutationMutation, AddUserToRolesMutationMutationVariables>(AddUserToRolesMutationDocument, options);
      }
export type AddUserToRolesMutationMutationHookResult = ReturnType<typeof useAddUserToRolesMutationMutation>;
export type AddUserToRolesMutationMutationResult = Apollo.MutationResult<AddUserToRolesMutationMutation>;
export type AddUserToRolesMutationMutationOptions = Apollo.BaseMutationOptions<AddUserToRolesMutationMutation, AddUserToRolesMutationMutationVariables>;