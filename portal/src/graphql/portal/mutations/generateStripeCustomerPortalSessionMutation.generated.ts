import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GenerateStripeCustomerPortalSessionMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
}>;


export type GenerateStripeCustomerPortalSessionMutationMutation = { __typename?: 'Mutation', generateStripeCustomerPortalSession: { __typename?: 'GenerateStripeCustomerPortalSessionPayload', url: string } };


export const GenerateStripeCustomerPortalSessionMutationDocument = gql`
    mutation generateStripeCustomerPortalSessionMutation($appID: ID!) {
  generateStripeCustomerPortalSession(input: {appID: $appID}) {
    url
  }
}
    `;
export type GenerateStripeCustomerPortalSessionMutationMutationFn = Apollo.MutationFunction<GenerateStripeCustomerPortalSessionMutationMutation, GenerateStripeCustomerPortalSessionMutationMutationVariables>;

/**
 * __useGenerateStripeCustomerPortalSessionMutationMutation__
 *
 * To run a mutation, you first call `useGenerateStripeCustomerPortalSessionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useGenerateStripeCustomerPortalSessionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [generateStripeCustomerPortalSessionMutationMutation, { data, loading, error }] = useGenerateStripeCustomerPortalSessionMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *   },
 * });
 */
export function useGenerateStripeCustomerPortalSessionMutationMutation(baseOptions?: Apollo.MutationHookOptions<GenerateStripeCustomerPortalSessionMutationMutation, GenerateStripeCustomerPortalSessionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<GenerateStripeCustomerPortalSessionMutationMutation, GenerateStripeCustomerPortalSessionMutationMutationVariables>(GenerateStripeCustomerPortalSessionMutationDocument, options);
      }
export type GenerateStripeCustomerPortalSessionMutationMutationHookResult = ReturnType<typeof useGenerateStripeCustomerPortalSessionMutationMutation>;
export type GenerateStripeCustomerPortalSessionMutationMutationResult = Apollo.MutationResult<GenerateStripeCustomerPortalSessionMutationMutation>;
export type GenerateStripeCustomerPortalSessionMutationMutationOptions = Apollo.BaseMutationOptions<GenerateStripeCustomerPortalSessionMutationMutation, GenerateStripeCustomerPortalSessionMutationMutationVariables>;