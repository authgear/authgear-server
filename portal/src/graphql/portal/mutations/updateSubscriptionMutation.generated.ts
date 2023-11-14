import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateSubscriptionMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  planName: Types.Scalars['String']['input'];
}>;


export type UpdateSubscriptionMutationMutation = { __typename?: 'Mutation', updateSubscription: { __typename?: 'UpdateSubscriptionPayload', app: { __typename?: 'App', id: string, planName: string } } };


export const UpdateSubscriptionMutationDocument = gql`
    mutation updateSubscriptionMutation($appID: ID!, $planName: String!) {
  updateSubscription(input: {appID: $appID, planName: $planName}) {
    app {
      id
      planName
    }
  }
}
    `;
export type UpdateSubscriptionMutationMutationFn = Apollo.MutationFunction<UpdateSubscriptionMutationMutation, UpdateSubscriptionMutationMutationVariables>;

/**
 * __useUpdateSubscriptionMutationMutation__
 *
 * To run a mutation, you first call `useUpdateSubscriptionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateSubscriptionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateSubscriptionMutationMutation, { data, loading, error }] = useUpdateSubscriptionMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      planName: // value for 'planName'
 *   },
 * });
 */
export function useUpdateSubscriptionMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateSubscriptionMutationMutation, UpdateSubscriptionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateSubscriptionMutationMutation, UpdateSubscriptionMutationMutationVariables>(UpdateSubscriptionMutationDocument, options);
      }
export type UpdateSubscriptionMutationMutationHookResult = ReturnType<typeof useUpdateSubscriptionMutationMutation>;
export type UpdateSubscriptionMutationMutationResult = Apollo.MutationResult<UpdateSubscriptionMutationMutation>;
export type UpdateSubscriptionMutationMutationOptions = Apollo.BaseMutationOptions<UpdateSubscriptionMutationMutation, UpdateSubscriptionMutationMutationVariables>;