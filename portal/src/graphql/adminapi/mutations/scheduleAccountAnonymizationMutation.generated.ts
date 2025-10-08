import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ScheduleAccountAnonymizationMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type ScheduleAccountAnonymizationMutationMutation = { __typename?: 'Mutation', scheduleAccountAnonymization: { __typename?: 'ScheduleAccountAnonymizationPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, temporarilyDisabledFrom?: any | null, temporarilyDisabledUntil?: any | null, accountValidFrom?: any | null, accountValidUntil?: any | null } } };


export const ScheduleAccountAnonymizationMutationDocument = gql`
    mutation scheduleAccountAnonymizationMutation($userID: ID!) {
  scheduleAccountAnonymization(input: {userID: $userID}) {
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
export type ScheduleAccountAnonymizationMutationMutationFn = Apollo.MutationFunction<ScheduleAccountAnonymizationMutationMutation, ScheduleAccountAnonymizationMutationMutationVariables>;

/**
 * __useScheduleAccountAnonymizationMutationMutation__
 *
 * To run a mutation, you first call `useScheduleAccountAnonymizationMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useScheduleAccountAnonymizationMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [scheduleAccountAnonymizationMutationMutation, { data, loading, error }] = useScheduleAccountAnonymizationMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useScheduleAccountAnonymizationMutationMutation(baseOptions?: Apollo.MutationHookOptions<ScheduleAccountAnonymizationMutationMutation, ScheduleAccountAnonymizationMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ScheduleAccountAnonymizationMutationMutation, ScheduleAccountAnonymizationMutationMutationVariables>(ScheduleAccountAnonymizationMutationDocument, options);
      }
export type ScheduleAccountAnonymizationMutationMutationHookResult = ReturnType<typeof useScheduleAccountAnonymizationMutationMutation>;
export type ScheduleAccountAnonymizationMutationMutationResult = Apollo.MutationResult<ScheduleAccountAnonymizationMutationMutation>;
export type ScheduleAccountAnonymizationMutationMutationOptions = Apollo.BaseMutationOptions<ScheduleAccountAnonymizationMutationMutation, ScheduleAccountAnonymizationMutationMutationVariables>;