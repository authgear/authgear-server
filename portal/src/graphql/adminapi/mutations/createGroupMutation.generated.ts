import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateGroupMutationMutationVariables = Types.Exact<{
  key: Types.Scalars['String']['input'];
  name: Types.Scalars['String']['input'];
  description?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type CreateGroupMutationMutation = { __typename?: 'Mutation', createGroup: { __typename?: 'CreateGroupPayload', group: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } } };


export const CreateGroupMutationDocument = gql`
    mutation createGroupMutation($key: String!, $name: String!, $description: String) {
  createGroup(input: {key: $key, name: $name, description: $description}) {
    group {
      id
      key
      name
      description
    }
  }
}
    `;
export type CreateGroupMutationMutationFn = Apollo.MutationFunction<CreateGroupMutationMutation, CreateGroupMutationMutationVariables>;

/**
 * __useCreateGroupMutationMutation__
 *
 * To run a mutation, you first call `useCreateGroupMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateGroupMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createGroupMutationMutation, { data, loading, error }] = useCreateGroupMutationMutation({
 *   variables: {
 *      key: // value for 'key'
 *      name: // value for 'name'
 *      description: // value for 'description'
 *   },
 * });
 */
export function useCreateGroupMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateGroupMutationMutation, CreateGroupMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateGroupMutationMutation, CreateGroupMutationMutationVariables>(CreateGroupMutationDocument, options);
      }
export type CreateGroupMutationMutationHookResult = ReturnType<typeof useCreateGroupMutationMutation>;
export type CreateGroupMutationMutationResult = Apollo.MutationResult<CreateGroupMutationMutation>;
export type CreateGroupMutationMutationOptions = Apollo.BaseMutationOptions<CreateGroupMutationMutation, CreateGroupMutationMutationVariables>;