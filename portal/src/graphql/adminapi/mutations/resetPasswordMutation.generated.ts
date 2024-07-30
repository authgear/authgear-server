import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ResetPasswordMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  password: Types.Scalars['String']['input'];
  sendPassword?: Types.InputMaybe<Types.Scalars['Boolean']['input']>;
  forceChangeOnLogin?: Types.InputMaybe<Types.Scalars['Boolean']['input']>;
}>;


export type ResetPasswordMutationMutation = { __typename?: 'Mutation', resetPassword: { __typename?: 'ResetPasswordPayload', user: { __typename?: 'User', id: string } } };


export const ResetPasswordMutationDocument = gql`
    mutation resetPasswordMutation($userID: ID!, $password: String!, $sendPassword: Boolean, $forceChangeOnLogin: Boolean) {
  resetPassword(
    input: {userID: $userID, password: $password, sendPassword: $sendPassword, forceChangeOnLogin: $forceChangeOnLogin}
  ) {
    user {
      id
    }
  }
}
    `;
export type ResetPasswordMutationMutationFn = Apollo.MutationFunction<ResetPasswordMutationMutation, ResetPasswordMutationMutationVariables>;

/**
 * __useResetPasswordMutationMutation__
 *
 * To run a mutation, you first call `useResetPasswordMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useResetPasswordMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [resetPasswordMutationMutation, { data, loading, error }] = useResetPasswordMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      password: // value for 'password'
 *      sendPassword: // value for 'sendPassword'
 *      forceChangeOnLogin: // value for 'forceChangeOnLogin'
 *   },
 * });
 */
export function useResetPasswordMutationMutation(baseOptions?: Apollo.MutationHookOptions<ResetPasswordMutationMutation, ResetPasswordMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ResetPasswordMutationMutation, ResetPasswordMutationMutationVariables>(ResetPasswordMutationDocument, options);
      }
export type ResetPasswordMutationMutationHookResult = ReturnType<typeof useResetPasswordMutationMutation>;
export type ResetPasswordMutationMutationResult = Apollo.MutationResult<ResetPasswordMutationMutation>;
export type ResetPasswordMutationMutationOptions = Apollo.BaseMutationOptions<ResetPasswordMutationMutation, ResetPasswordMutationMutationVariables>;