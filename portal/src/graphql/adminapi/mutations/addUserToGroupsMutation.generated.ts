import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AddUserToGroupsMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  groupKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type AddUserToGroupsMutationMutation = { __typename?: 'Mutation', addUserToGroups: { __typename?: 'AddUserToGroupsPayload', user: { __typename?: 'User', id: string, formattedName?: string | null } } };


export const AddUserToGroupsMutationDocument = gql`
    mutation addUserToGroupsMutation($userID: ID!, $groupKeys: [String!]!) {
  addUserToGroups(input: {userID: $userID, groupKeys: $groupKeys}) {
    user {
      id
      formattedName
    }
  }
}
    `;
export type AddUserToGroupsMutationMutationFn = Apollo.MutationFunction<AddUserToGroupsMutationMutation, AddUserToGroupsMutationMutationVariables>;

/**
 * __useAddUserToGroupsMutationMutation__
 *
 * To run a mutation, you first call `useAddUserToGroupsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddUserToGroupsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addUserToGroupsMutationMutation, { data, loading, error }] = useAddUserToGroupsMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      groupKeys: // value for 'groupKeys'
 *   },
 * });
 */
export function useAddUserToGroupsMutationMutation(baseOptions?: Apollo.MutationHookOptions<AddUserToGroupsMutationMutation, AddUserToGroupsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddUserToGroupsMutationMutation, AddUserToGroupsMutationMutationVariables>(AddUserToGroupsMutationDocument, options);
      }
export type AddUserToGroupsMutationMutationHookResult = ReturnType<typeof useAddUserToGroupsMutationMutation>;
export type AddUserToGroupsMutationMutationResult = Apollo.MutationResult<AddUserToGroupsMutationMutation>;
export type AddUserToGroupsMutationMutationOptions = Apollo.BaseMutationOptions<AddUserToGroupsMutationMutation, AddUserToGroupsMutationMutationVariables>;