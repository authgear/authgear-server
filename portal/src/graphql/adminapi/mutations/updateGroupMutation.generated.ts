import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateGroupMutationMutationVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
  key?: Types.InputMaybe<Types.Scalars['String']['input']>;
  name?: Types.InputMaybe<Types.Scalars['String']['input']>;
  description?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type UpdateGroupMutationMutation = { __typename?: 'Mutation', updateGroup: { __typename?: 'UpdateGroupPayload', group: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } } };


export const UpdateGroupMutationDocument = gql`
    mutation updateGroupMutation($id: ID!, $key: String, $name: String, $description: String) {
  updateGroup(input: {id: $id, key: $key, name: $name, description: $description}) {
    group {
      id
      key
      name
      description
    }
  }
}
    `;
export type UpdateGroupMutationMutationFn = Apollo.MutationFunction<UpdateGroupMutationMutation, UpdateGroupMutationMutationVariables>;

/**
 * __useUpdateGroupMutationMutation__
 *
 * To run a mutation, you first call `useUpdateGroupMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateGroupMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateGroupMutationMutation, { data, loading, error }] = useUpdateGroupMutationMutation({
 *   variables: {
 *      id: // value for 'id'
 *      key: // value for 'key'
 *      name: // value for 'name'
 *      description: // value for 'description'
 *   },
 * });
 */
export function useUpdateGroupMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateGroupMutationMutation, UpdateGroupMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateGroupMutationMutation, UpdateGroupMutationMutationVariables>(UpdateGroupMutationDocument, options);
      }
export type UpdateGroupMutationMutationHookResult = ReturnType<typeof useUpdateGroupMutationMutation>;
export type UpdateGroupMutationMutationResult = Apollo.MutationResult<UpdateGroupMutationMutation>;
export type UpdateGroupMutationMutationOptions = Apollo.BaseMutationOptions<UpdateGroupMutationMutation, UpdateGroupMutationMutationVariables>;