import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateRoleMutationMutationVariables = Types.Exact<{
  key: Types.Scalars['String']['input'];
  name: Types.Scalars['String']['input'];
  description?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type CreateRoleMutationMutation = { __typename?: 'Mutation', createRole: { __typename?: 'CreateRolePayload', role: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } } };


export const CreateRoleMutationDocument = gql`
    mutation createRoleMutation($key: String!, $name: String!, $description: String) {
  createRole(input: {key: $key, name: $name, description: $description}) {
    role {
      id
      key
      name
      description
    }
  }
}
    `;
export type CreateRoleMutationMutationFn = Apollo.MutationFunction<CreateRoleMutationMutation, CreateRoleMutationMutationVariables>;

/**
 * __useCreateRoleMutationMutation__
 *
 * To run a mutation, you first call `useCreateRoleMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateRoleMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createRoleMutationMutation, { data, loading, error }] = useCreateRoleMutationMutation({
 *   variables: {
 *      key: // value for 'key'
 *      name: // value for 'name'
 *      description: // value for 'description'
 *   },
 * });
 */
export function useCreateRoleMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateRoleMutationMutation, CreateRoleMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateRoleMutationMutation, CreateRoleMutationMutationVariables>(CreateRoleMutationDocument, options);
      }
export type CreateRoleMutationMutationHookResult = ReturnType<typeof useCreateRoleMutationMutation>;
export type CreateRoleMutationMutationResult = Apollo.MutationResult<CreateRoleMutationMutation>;
export type CreateRoleMutationMutationOptions = Apollo.BaseMutationOptions<CreateRoleMutationMutation, CreateRoleMutationMutationVariables>;