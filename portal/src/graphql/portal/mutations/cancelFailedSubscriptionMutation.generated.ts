import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CancelFailedSubscriptionMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
}>;


export type CancelFailedSubscriptionMutationMutation = { __typename?: 'Mutation', cancelFailedSubscription: { __typename?: 'CancelFailedSubscriptionPayload', app: { __typename?: 'App', id: string, isProcessingSubscription: boolean, lastStripeError?: any | null } } };


export const CancelFailedSubscriptionMutationDocument = gql`
    mutation cancelFailedSubscriptionMutation($appID: ID!) {
  cancelFailedSubscription(input: {appID: $appID}) {
    app {
      id
      isProcessingSubscription
      lastStripeError
    }
  }
}
    `;
export type CancelFailedSubscriptionMutationMutationFn = Apollo.MutationFunction<CancelFailedSubscriptionMutationMutation, CancelFailedSubscriptionMutationMutationVariables>;

/**
 * __useCancelFailedSubscriptionMutationMutation__
 *
 * To run a mutation, you first call `useCancelFailedSubscriptionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCancelFailedSubscriptionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [cancelFailedSubscriptionMutationMutation, { data, loading, error }] = useCancelFailedSubscriptionMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *   },
 * });
 */
export function useCancelFailedSubscriptionMutationMutation(baseOptions?: Apollo.MutationHookOptions<CancelFailedSubscriptionMutationMutation, CancelFailedSubscriptionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CancelFailedSubscriptionMutationMutation, CancelFailedSubscriptionMutationMutationVariables>(CancelFailedSubscriptionMutationDocument, options);
      }
export type CancelFailedSubscriptionMutationMutationHookResult = ReturnType<typeof useCancelFailedSubscriptionMutationMutation>;
export type CancelFailedSubscriptionMutationMutationResult = Apollo.MutationResult<CancelFailedSubscriptionMutationMutation>;
export type CancelFailedSubscriptionMutationMutationOptions = Apollo.BaseMutationOptions<CancelFailedSubscriptionMutationMutation, CancelFailedSubscriptionMutationMutationVariables>;