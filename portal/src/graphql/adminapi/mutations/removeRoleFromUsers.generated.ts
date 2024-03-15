import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveRoleFromUsersMutationMutationVariables = Types.Exact<{
  roleKey: Types.Scalars['String']['input'];
  userIDs: Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input'];
}>;


export type RemoveRoleFromUsersMutationMutation = { __typename?: 'Mutation', removeRoleFromUsers: { __typename?: 'RemoveRoleFromUsersPayload', role: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } } };


export const RemoveRoleFromUsersMutationDocument = gql`
    mutation removeRoleFromUsersMutation($roleKey: String!, $userIDs: [ID!]!) {
  removeRoleFromUsers(input: {roleKey: $roleKey, userIDs: $userIDs}) {
    role {
      id
      key
      name
      description
    }
  }
}
    `;
export type RemoveRoleFromUsersMutationMutationFn = Apollo.MutationFunction<RemoveRoleFromUsersMutationMutation, RemoveRoleFromUsersMutationMutationVariables>;

/**
 * __useRemoveRoleFromUsersMutationMutation__
 *
 * To run a mutation, you first call `useRemoveRoleFromUsersMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveRoleFromUsersMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeRoleFromUsersMutationMutation, { data, loading, error }] = useRemoveRoleFromUsersMutationMutation({
 *   variables: {
 *      roleKey: // value for 'roleKey'
 *      userIDs: // value for 'userIDs'
 *   },
 * });
 */
export function useRemoveRoleFromUsersMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveRoleFromUsersMutationMutation, RemoveRoleFromUsersMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveRoleFromUsersMutationMutation, RemoveRoleFromUsersMutationMutationVariables>(RemoveRoleFromUsersMutationDocument, options);
      }
export type RemoveRoleFromUsersMutationMutationHookResult = ReturnType<typeof useRemoveRoleFromUsersMutationMutation>;
export type RemoveRoleFromUsersMutationMutationResult = Apollo.MutationResult<RemoveRoleFromUsersMutationMutation>;
export type RemoveRoleFromUsersMutationMutationOptions = Apollo.BaseMutationOptions<RemoveRoleFromUsersMutationMutation, RemoveRoleFromUsersMutationMutationVariables>;