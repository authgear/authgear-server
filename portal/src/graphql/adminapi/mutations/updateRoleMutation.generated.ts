import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateRoleMutationMutationVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
  key?: Types.InputMaybe<Types.Scalars['String']['input']>;
  name?: Types.InputMaybe<Types.Scalars['String']['input']>;
  description?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type UpdateRoleMutationMutation = { __typename?: 'Mutation', updateRole: { __typename?: 'UpdateRolePayload', role: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } } };


export const UpdateRoleMutationDocument = gql`
    mutation updateRoleMutation($id: ID!, $key: String, $name: String, $description: String) {
  updateRole(input: {id: $id, key: $key, name: $name, description: $description}) {
    role {
      id
      key
      name
      description
    }
  }
}
    `;
export type UpdateRoleMutationMutationFn = Apollo.MutationFunction<UpdateRoleMutationMutation, UpdateRoleMutationMutationVariables>;

/**
 * __useUpdateRoleMutationMutation__
 *
 * To run a mutation, you first call `useUpdateRoleMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateRoleMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateRoleMutationMutation, { data, loading, error }] = useUpdateRoleMutationMutation({
 *   variables: {
 *      id: // value for 'id'
 *      key: // value for 'key'
 *      name: // value for 'name'
 *      description: // value for 'description'
 *   },
 * });
 */
export function useUpdateRoleMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateRoleMutationMutation, UpdateRoleMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateRoleMutationMutation, UpdateRoleMutationMutationVariables>(UpdateRoleMutationDocument, options);
      }
export type UpdateRoleMutationMutationHookResult = ReturnType<typeof useUpdateRoleMutationMutation>;
export type UpdateRoleMutationMutationResult = Apollo.MutationResult<UpdateRoleMutationMutation>;
export type UpdateRoleMutationMutationOptions = Apollo.BaseMutationOptions<UpdateRoleMutationMutation, UpdateRoleMutationMutationVariables>;