import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetPasswordExpiredMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  isExpired: Types.Scalars['Boolean']['input'];
}>;


export type SetPasswordExpiredMutation = { __typename?: 'Mutation', setPasswordExpired: { __typename?: 'SetPasswordExpiredPayload', user: { __typename?: 'User', id: string } } };


export const SetPasswordExpiredDocument = gql`
    mutation setPasswordExpired($userID: ID!, $isExpired: Boolean!) {
  setPasswordExpired(input: {userID: $userID, isExpired: $isExpired}) {
    user {
      id
    }
  }
}
    `;
export type SetPasswordExpiredMutationFn = Apollo.MutationFunction<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>;

/**
 * __useSetPasswordExpiredMutation__
 *
 * To run a mutation, you first call `useSetPasswordExpiredMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetPasswordExpiredMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setPasswordExpiredMutation, { data, loading, error }] = useSetPasswordExpiredMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      isExpired: // value for 'isExpired'
 *   },
 * });
 */
export function useSetPasswordExpiredMutation(baseOptions?: Apollo.MutationHookOptions<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>(SetPasswordExpiredDocument, options);
      }
export type SetPasswordExpiredMutationHookResult = ReturnType<typeof useSetPasswordExpiredMutation>;
export type SetPasswordExpiredMutationResult = Apollo.MutationResult<SetPasswordExpiredMutation>;
export type SetPasswordExpiredMutationOptions = Apollo.BaseMutationOptions<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>;