import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ResetAccountLockoutMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type ResetAccountLockoutMutationMutation = { __typename?: 'Mutation', resetAccountLockout: { __typename?: 'ResetAccountLockoutPayload', user: { __typename?: 'User', id: string, accountLockout: { __typename?: 'AccountLockout', isLocked: boolean, lockoutType: string, lockedUntil?: any | null, lockedIPs: Array<{ __typename?: 'LockedIP', ipAddress: string, lockedUntil: any }> } } } };


export const ResetAccountLockoutMutationDocument = gql`
    mutation resetAccountLockoutMutation($userID: ID!) {
  resetAccountLockout(input: {userID: $userID}) {
    user {
      id
      accountLockout {
        isLocked
        lockoutType
        lockedUntil
        lockedIPs {
          ipAddress
          lockedUntil
        }
      }
    }
  }
}
    `;
export type ResetAccountLockoutMutationMutationFn = Apollo.MutationFunction<ResetAccountLockoutMutationMutation, ResetAccountLockoutMutationMutationVariables>;

/**
 * __useResetAccountLockoutMutationMutation__
 *
 * To run a mutation, you first call `useResetAccountLockoutMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useResetAccountLockoutMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [resetAccountLockoutMutationMutation, { data, loading, error }] = useResetAccountLockoutMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useResetAccountLockoutMutationMutation(baseOptions?: Apollo.MutationHookOptions<ResetAccountLockoutMutationMutation, ResetAccountLockoutMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ResetAccountLockoutMutationMutation, ResetAccountLockoutMutationMutationVariables>(ResetAccountLockoutMutationDocument, options);
      }
export type ResetAccountLockoutMutationMutationHookResult = ReturnType<typeof useResetAccountLockoutMutationMutation>;
export type ResetAccountLockoutMutationMutationResult = Apollo.MutationResult<ResetAccountLockoutMutationMutation>;
export type ResetAccountLockoutMutationMutationOptions = Apollo.BaseMutationOptions<ResetAccountLockoutMutationMutation, ResetAccountLockoutMutationMutationVariables>;