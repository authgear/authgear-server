import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type PreviewUpdateSubscriptionMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  planName: Types.Scalars['String']['input'];
}>;


export type PreviewUpdateSubscriptionMutationMutation = { __typename?: 'Mutation', previewUpdateSubscription: { __typename?: 'PreviewUpdateSubscriptionPayload', currency: string, amountDue: number } };


export const PreviewUpdateSubscriptionMutationDocument = gql`
    mutation previewUpdateSubscriptionMutation($appID: ID!, $planName: String!) {
  previewUpdateSubscription(input: {appID: $appID, planName: $planName}) {
    currency
    amountDue
  }
}
    `;
export type PreviewUpdateSubscriptionMutationMutationFn = Apollo.MutationFunction<PreviewUpdateSubscriptionMutationMutation, PreviewUpdateSubscriptionMutationMutationVariables>;

/**
 * __usePreviewUpdateSubscriptionMutationMutation__
 *
 * To run a mutation, you first call `usePreviewUpdateSubscriptionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `usePreviewUpdateSubscriptionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [previewUpdateSubscriptionMutationMutation, { data, loading, error }] = usePreviewUpdateSubscriptionMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      planName: // value for 'planName'
 *   },
 * });
 */
export function usePreviewUpdateSubscriptionMutationMutation(baseOptions?: Apollo.MutationHookOptions<PreviewUpdateSubscriptionMutationMutation, PreviewUpdateSubscriptionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<PreviewUpdateSubscriptionMutationMutation, PreviewUpdateSubscriptionMutationMutationVariables>(PreviewUpdateSubscriptionMutationDocument, options);
      }
export type PreviewUpdateSubscriptionMutationMutationHookResult = ReturnType<typeof usePreviewUpdateSubscriptionMutationMutation>;
export type PreviewUpdateSubscriptionMutationMutationResult = Apollo.MutationResult<PreviewUpdateSubscriptionMutationMutation>;
export type PreviewUpdateSubscriptionMutationMutationOptions = Apollo.BaseMutationOptions<PreviewUpdateSubscriptionMutationMutation, PreviewUpdateSubscriptionMutationMutationVariables>;