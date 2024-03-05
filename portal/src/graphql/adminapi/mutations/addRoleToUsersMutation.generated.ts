import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddRoleToUsersMutationMutationVariables = Types.Exact<{
  roleKey: Types.Scalars['String']['input'];
  userIDs: Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input'];
}>;


export type AddRoleToUsersMutationMutation = { __typename?: 'Mutation', addRoleToUsers: { __typename?: 'AddRoleToUsersPayload', role: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } } };


export const AddRoleToUsersMutationDocument = gql`
    mutation addRoleToUsersMutation($roleKey: String!, $userIDs: [ID!]!) {
  addRoleToUsers(input: {roleKey: $roleKey, userIDs: $userIDs}) {
    role {
      id
      key
      name
      description
    }
  }
}
    `;
export type AddRoleToUsersMutationMutationFn = Apollo.MutationFunction<AddRoleToUsersMutationMutation, AddRoleToUsersMutationMutationVariables>;

/**
 * __useAddRoleToUsersMutationMutation__
 *
 * To run a mutation, you first call `useAddRoleToUsersMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddRoleToUsersMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addRoleToUsersMutationMutation, { data, loading, error }] = useAddRoleToUsersMutationMutation({
 *   variables: {
 *      roleKey: // value for 'roleKey'
 *      userIDs: // value for 'userIDs'
 *   },
 * });
 */
export function useAddRoleToUsersMutationMutation(baseOptions?: Apollo.MutationHookOptions<AddRoleToUsersMutationMutation, AddRoleToUsersMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddRoleToUsersMutationMutation, AddRoleToUsersMutationMutationVariables>(AddRoleToUsersMutationDocument, options);
      }
export type AddRoleToUsersMutationMutationHookResult = ReturnType<typeof useAddRoleToUsersMutationMutation>;
export type AddRoleToUsersMutationMutationResult = Apollo.MutationResult<AddRoleToUsersMutationMutation>;
export type AddRoleToUsersMutationMutationOptions = Apollo.BaseMutationOptions<AddRoleToUsersMutationMutation, AddRoleToUsersMutationMutationVariables>;