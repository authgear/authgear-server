import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ScheduleAccountDeletionMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type ScheduleAccountDeletionMutationMutation = { __typename?: 'Mutation', scheduleAccountDeletion: { __typename?: 'ScheduleAccountDeletionPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null } } };


export const ScheduleAccountDeletionMutationDocument = gql`
    mutation scheduleAccountDeletionMutation($userID: ID!) {
  scheduleAccountDeletion(input: {userID: $userID}) {
    user {
      id
      isDisabled
      disableReason
      isDeactivated
      deleteAt
      isAnonymized
      anonymizeAt
    }
  }
}
    `;
export type ScheduleAccountDeletionMutationMutationFn = Apollo.MutationFunction<ScheduleAccountDeletionMutationMutation, ScheduleAccountDeletionMutationMutationVariables>;

/**
 * __useScheduleAccountDeletionMutationMutation__
 *
 * To run a mutation, you first call `useScheduleAccountDeletionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useScheduleAccountDeletionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [scheduleAccountDeletionMutationMutation, { data, loading, error }] = useScheduleAccountDeletionMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useScheduleAccountDeletionMutationMutation(baseOptions?: Apollo.MutationHookOptions<ScheduleAccountDeletionMutationMutation, ScheduleAccountDeletionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ScheduleAccountDeletionMutationMutation, ScheduleAccountDeletionMutationMutationVariables>(ScheduleAccountDeletionMutationDocument, options);
      }
export type ScheduleAccountDeletionMutationMutationHookResult = ReturnType<typeof useScheduleAccountDeletionMutationMutation>;
export type ScheduleAccountDeletionMutationMutationResult = Apollo.MutationResult<ScheduleAccountDeletionMutationMutation>;
export type ScheduleAccountDeletionMutationMutationOptions = Apollo.BaseMutationOptions<ScheduleAccountDeletionMutationMutation, ScheduleAccountDeletionMutationMutationVariables>;