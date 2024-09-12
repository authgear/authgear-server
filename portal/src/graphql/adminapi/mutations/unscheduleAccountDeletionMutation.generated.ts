import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UnscheduleAccountDeletionMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type UnscheduleAccountDeletionMutationMutation = { __typename?: 'Mutation', unscheduleAccountDeletion: { __typename?: 'UnscheduleAccountDeletionPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null } } };


export const UnscheduleAccountDeletionMutationDocument = gql`
    mutation unscheduleAccountDeletionMutation($userID: ID!) {
  unscheduleAccountDeletion(input: {userID: $userID}) {
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
export type UnscheduleAccountDeletionMutationMutationFn = Apollo.MutationFunction<UnscheduleAccountDeletionMutationMutation, UnscheduleAccountDeletionMutationMutationVariables>;

/**
 * __useUnscheduleAccountDeletionMutationMutation__
 *
 * To run a mutation, you first call `useUnscheduleAccountDeletionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUnscheduleAccountDeletionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [unscheduleAccountDeletionMutationMutation, { data, loading, error }] = useUnscheduleAccountDeletionMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useUnscheduleAccountDeletionMutationMutation(baseOptions?: Apollo.MutationHookOptions<UnscheduleAccountDeletionMutationMutation, UnscheduleAccountDeletionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UnscheduleAccountDeletionMutationMutation, UnscheduleAccountDeletionMutationMutationVariables>(UnscheduleAccountDeletionMutationDocument, options);
      }
export type UnscheduleAccountDeletionMutationMutationHookResult = ReturnType<typeof useUnscheduleAccountDeletionMutationMutation>;
export type UnscheduleAccountDeletionMutationMutationResult = Apollo.MutationResult<UnscheduleAccountDeletionMutationMutation>;
export type UnscheduleAccountDeletionMutationMutationOptions = Apollo.BaseMutationOptions<UnscheduleAccountDeletionMutationMutation, UnscheduleAccountDeletionMutationMutationVariables>;