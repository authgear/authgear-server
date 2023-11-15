import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateCheckoutSessionMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  planName: Types.Scalars['String']['input'];
}>;


export type CreateCheckoutSessionMutationMutation = { __typename?: 'Mutation', createCheckoutSession: { __typename?: 'CreateCheckoutSessionPayload', url: string } };


export const CreateCheckoutSessionMutationDocument = gql`
    mutation createCheckoutSessionMutation($appID: ID!, $planName: String!) {
  createCheckoutSession(input: {appID: $appID, planName: $planName}) {
    url
  }
}
    `;
export type CreateCheckoutSessionMutationMutationFn = Apollo.MutationFunction<CreateCheckoutSessionMutationMutation, CreateCheckoutSessionMutationMutationVariables>;

/**
 * __useCreateCheckoutSessionMutationMutation__
 *
 * To run a mutation, you first call `useCreateCheckoutSessionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateCheckoutSessionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createCheckoutSessionMutationMutation, { data, loading, error }] = useCreateCheckoutSessionMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      planName: // value for 'planName'
 *   },
 * });
 */
export function useCreateCheckoutSessionMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateCheckoutSessionMutationMutation, CreateCheckoutSessionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateCheckoutSessionMutationMutation, CreateCheckoutSessionMutationMutationVariables>(CreateCheckoutSessionMutationDocument, options);
      }
export type CreateCheckoutSessionMutationMutationHookResult = ReturnType<typeof useCreateCheckoutSessionMutationMutation>;
export type CreateCheckoutSessionMutationMutationResult = Apollo.MutationResult<CreateCheckoutSessionMutationMutation>;
export type CreateCheckoutSessionMutationMutationOptions = Apollo.BaseMutationOptions<CreateCheckoutSessionMutationMutation, CreateCheckoutSessionMutationMutationVariables>;