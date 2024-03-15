import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveUserFromGroupsMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  groupKeys: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type RemoveUserFromGroupsMutationMutation = { __typename?: 'Mutation', removeUserFromGroups: { __typename?: 'RemoveUserFromGroupsPayload', user: { __typename?: 'User', id: string, formattedName?: string | null } } };


export const RemoveUserFromGroupsMutationDocument = gql`
    mutation removeUserFromGroupsMutation($userID: ID!, $groupKeys: [String!]!) {
  removeUserFromGroups(input: {userID: $userID, groupKeys: $groupKeys}) {
    user {
      id
      formattedName
    }
  }
}
    `;
export type RemoveUserFromGroupsMutationMutationFn = Apollo.MutationFunction<RemoveUserFromGroupsMutationMutation, RemoveUserFromGroupsMutationMutationVariables>;

/**
 * __useRemoveUserFromGroupsMutationMutation__
 *
 * To run a mutation, you first call `useRemoveUserFromGroupsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveUserFromGroupsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeUserFromGroupsMutationMutation, { data, loading, error }] = useRemoveUserFromGroupsMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      groupKeys: // value for 'groupKeys'
 *   },
 * });
 */
export function useRemoveUserFromGroupsMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveUserFromGroupsMutationMutation, RemoveUserFromGroupsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveUserFromGroupsMutationMutation, RemoveUserFromGroupsMutationMutationVariables>(RemoveUserFromGroupsMutationDocument, options);
      }
export type RemoveUserFromGroupsMutationMutationHookResult = ReturnType<typeof useRemoveUserFromGroupsMutationMutation>;
export type RemoveUserFromGroupsMutationMutationResult = Apollo.MutationResult<RemoveUserFromGroupsMutationMutation>;
export type RemoveUserFromGroupsMutationMutationOptions = Apollo.BaseMutationOptions<RemoveUserFromGroupsMutationMutation, RemoveUserFromGroupsMutationMutationVariables>;