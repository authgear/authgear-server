import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UnscheduleAccountAnonymizationMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type UnscheduleAccountAnonymizationMutationMutation = { __typename?: 'Mutation', unscheduleAccountAnonymization: { __typename?: 'UnscheduleAccountAnonymizationPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, temporarilyDisabledFrom?: any | null, temporarilyDisabledUntil?: any | null, accountValidFrom?: any | null, accountValidUntil?: any | null } } };


export const UnscheduleAccountAnonymizationMutationDocument = gql`
    mutation unscheduleAccountAnonymizationMutation($userID: ID!) {
  unscheduleAccountAnonymization(input: {userID: $userID}) {
    user {
      id
      isDisabled
      disableReason
      isDeactivated
      deleteAt
      isAnonymized
      anonymizeAt
      temporarilyDisabledFrom
      temporarilyDisabledUntil
      accountValidFrom
      accountValidUntil
    }
  }
}
    `;
export type UnscheduleAccountAnonymizationMutationMutationFn = Apollo.MutationFunction<UnscheduleAccountAnonymizationMutationMutation, UnscheduleAccountAnonymizationMutationMutationVariables>;

/**
 * __useUnscheduleAccountAnonymizationMutationMutation__
 *
 * To run a mutation, you first call `useUnscheduleAccountAnonymizationMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUnscheduleAccountAnonymizationMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [unscheduleAccountAnonymizationMutationMutation, { data, loading, error }] = useUnscheduleAccountAnonymizationMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useUnscheduleAccountAnonymizationMutationMutation(baseOptions?: Apollo.MutationHookOptions<UnscheduleAccountAnonymizationMutationMutation, UnscheduleAccountAnonymizationMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UnscheduleAccountAnonymizationMutationMutation, UnscheduleAccountAnonymizationMutationMutationVariables>(UnscheduleAccountAnonymizationMutationDocument, options);
      }
export type UnscheduleAccountAnonymizationMutationMutationHookResult = ReturnType<typeof useUnscheduleAccountAnonymizationMutationMutation>;
export type UnscheduleAccountAnonymizationMutationMutationResult = Apollo.MutationResult<UnscheduleAccountAnonymizationMutationMutation>;
export type UnscheduleAccountAnonymizationMutationMutationOptions = Apollo.BaseMutationOptions<UnscheduleAccountAnonymizationMutationMutation, UnscheduleAccountAnonymizationMutationMutationVariables>;