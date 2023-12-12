import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetSubscriptionCancelledStatusMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  cancelled: Types.Scalars['Boolean']['input'];
}>;


export type SetSubscriptionCancelledStatusMutationMutation = { __typename?: 'Mutation', setSubscriptionCancelledStatus: { __typename?: 'SetSubscriptionCancelledStatusPayload', app: { __typename?: 'App', id: string, subscription?: { __typename?: 'Subscription', id: string, endedAt?: any | null } | null } } };


export const SetSubscriptionCancelledStatusMutationDocument = gql`
    mutation setSubscriptionCancelledStatusMutation($appID: ID!, $cancelled: Boolean!) {
  setSubscriptionCancelledStatus(input: {appID: $appID, cancelled: $cancelled}) {
    app {
      id
      subscription {
        id
        endedAt
      }
    }
  }
}
    `;
export type SetSubscriptionCancelledStatusMutationMutationFn = Apollo.MutationFunction<SetSubscriptionCancelledStatusMutationMutation, SetSubscriptionCancelledStatusMutationMutationVariables>;

/**
 * __useSetSubscriptionCancelledStatusMutationMutation__
 *
 * To run a mutation, you first call `useSetSubscriptionCancelledStatusMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetSubscriptionCancelledStatusMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setSubscriptionCancelledStatusMutationMutation, { data, loading, error }] = useSetSubscriptionCancelledStatusMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      cancelled: // value for 'cancelled'
 *   },
 * });
 */
export function useSetSubscriptionCancelledStatusMutationMutation(baseOptions?: Apollo.MutationHookOptions<SetSubscriptionCancelledStatusMutationMutation, SetSubscriptionCancelledStatusMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetSubscriptionCancelledStatusMutationMutation, SetSubscriptionCancelledStatusMutationMutationVariables>(SetSubscriptionCancelledStatusMutationDocument, options);
      }
export type SetSubscriptionCancelledStatusMutationMutationHookResult = ReturnType<typeof useSetSubscriptionCancelledStatusMutationMutation>;
export type SetSubscriptionCancelledStatusMutationMutationResult = Apollo.MutationResult<SetSubscriptionCancelledStatusMutationMutation>;
export type SetSubscriptionCancelledStatusMutationMutationOptions = Apollo.BaseMutationOptions<SetSubscriptionCancelledStatusMutationMutation, SetSubscriptionCancelledStatusMutationMutationVariables>;