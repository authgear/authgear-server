import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ReconcileCheckoutSessionMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  checkoutSessionID: Types.Scalars['String']['input'];
}>;


export type ReconcileCheckoutSessionMutationMutation = { __typename?: 'Mutation', reconcileCheckoutSession: { __typename?: 'reconcileCheckoutSessionPayload', app: { __typename?: 'App', id: string } } };


export const ReconcileCheckoutSessionMutationDocument = gql`
    mutation reconcileCheckoutSessionMutation($appID: ID!, $checkoutSessionID: String!) {
  reconcileCheckoutSession(
    input: {appID: $appID, checkoutSessionID: $checkoutSessionID}
  ) {
    app {
      id
    }
  }
}
    `;
export type ReconcileCheckoutSessionMutationMutationFn = Apollo.MutationFunction<ReconcileCheckoutSessionMutationMutation, ReconcileCheckoutSessionMutationMutationVariables>;

/**
 * __useReconcileCheckoutSessionMutationMutation__
 *
 * To run a mutation, you first call `useReconcileCheckoutSessionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useReconcileCheckoutSessionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [reconcileCheckoutSessionMutationMutation, { data, loading, error }] = useReconcileCheckoutSessionMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      checkoutSessionID: // value for 'checkoutSessionID'
 *   },
 * });
 */
export function useReconcileCheckoutSessionMutationMutation(baseOptions?: Apollo.MutationHookOptions<ReconcileCheckoutSessionMutationMutation, ReconcileCheckoutSessionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ReconcileCheckoutSessionMutationMutation, ReconcileCheckoutSessionMutationMutationVariables>(ReconcileCheckoutSessionMutationDocument, options);
      }
export type ReconcileCheckoutSessionMutationMutationHookResult = ReturnType<typeof useReconcileCheckoutSessionMutationMutation>;
export type ReconcileCheckoutSessionMutationMutationResult = Apollo.MutationResult<ReconcileCheckoutSessionMutationMutation>;
export type ReconcileCheckoutSessionMutationMutationOptions = Apollo.BaseMutationOptions<ReconcileCheckoutSessionMutationMutation, ReconcileCheckoutSessionMutationMutationVariables>;