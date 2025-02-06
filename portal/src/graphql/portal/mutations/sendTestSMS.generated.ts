import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SendTestSmsMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  to: Types.Scalars['String']['input'];
  config: Types.SmsProviderConfigurationInput;
}>;


export type SendTestSmsMutationMutation = { __typename?: 'Mutation', sendTestSMSConfiguration?: boolean | null };


export const SendTestSmsMutationDocument = gql`
    mutation sendTestSMSMutation($appID: ID!, $to: String!, $config: SMSProviderConfigurationInput!) {
  sendTestSMSConfiguration(
    input: {appID: $appID, to: $to, providerConfiguration: $config}
  )
}
    `;
export type SendTestSmsMutationMutationFn = Apollo.MutationFunction<SendTestSmsMutationMutation, SendTestSmsMutationMutationVariables>;

/**
 * __useSendTestSmsMutationMutation__
 *
 * To run a mutation, you first call `useSendTestSmsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSendTestSmsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [sendTestSmsMutationMutation, { data, loading, error }] = useSendTestSmsMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      to: // value for 'to'
 *      config: // value for 'config'
 *   },
 * });
 */
export function useSendTestSmsMutationMutation(baseOptions?: Apollo.MutationHookOptions<SendTestSmsMutationMutation, SendTestSmsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SendTestSmsMutationMutation, SendTestSmsMutationMutationVariables>(SendTestSmsMutationDocument, options);
      }
export type SendTestSmsMutationMutationHookResult = ReturnType<typeof useSendTestSmsMutationMutation>;
export type SendTestSmsMutationMutationResult = Apollo.MutationResult<SendTestSmsMutationMutation>;
export type SendTestSmsMutationMutationOptions = Apollo.BaseMutationOptions<SendTestSmsMutationMutation, SendTestSmsMutationMutationVariables>;