import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateUserMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  standardAttributes: Types.Scalars['UserStandardAttributes']['input'];
  customAttributes: Types.Scalars['UserCustomAttributes']['input'];
}>;


export type UpdateUserMutationMutation = { __typename?: 'Mutation', updateUser: { __typename?: 'UpdateUserPayload', user: { __typename?: 'User', id: string, updatedAt: any, standardAttributes: any, customAttributes: any } } };


export const UpdateUserMutationDocument = gql`
    mutation updateUserMutation($userID: ID!, $standardAttributes: UserStandardAttributes!, $customAttributes: UserCustomAttributes!) {
  updateUser(
    input: {userID: $userID, standardAttributes: $standardAttributes, customAttributes: $customAttributes}
  ) {
    user {
      id
      updatedAt
      standardAttributes
      customAttributes
    }
  }
}
    `;
export type UpdateUserMutationMutationFn = Apollo.MutationFunction<UpdateUserMutationMutation, UpdateUserMutationMutationVariables>;

/**
 * __useUpdateUserMutationMutation__
 *
 * To run a mutation, you first call `useUpdateUserMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateUserMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateUserMutationMutation, { data, loading, error }] = useUpdateUserMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      standardAttributes: // value for 'standardAttributes'
 *      customAttributes: // value for 'customAttributes'
 *   },
 * });
 */
export function useUpdateUserMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateUserMutationMutation, UpdateUserMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateUserMutationMutation, UpdateUserMutationMutationVariables>(UpdateUserMutationDocument, options);
      }
export type UpdateUserMutationMutationHookResult = ReturnType<typeof useUpdateUserMutationMutation>;
export type UpdateUserMutationMutationResult = Apollo.MutationResult<UpdateUserMutationMutation>;
export type UpdateUserMutationMutationOptions = Apollo.BaseMutationOptions<UpdateUserMutationMutation, UpdateUserMutationMutationVariables>;