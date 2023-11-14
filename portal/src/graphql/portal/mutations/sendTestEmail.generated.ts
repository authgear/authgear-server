import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SendTestEmailMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  smtpHost: Types.Scalars['String']['input'];
  smtpPort: Types.Scalars['Int']['input'];
  smtpUsername: Types.Scalars['String']['input'];
  smtpPassword: Types.Scalars['String']['input'];
  to: Types.Scalars['String']['input'];
}>;


export type SendTestEmailMutationMutation = { __typename?: 'Mutation', sendTestSMTPConfigurationEmail?: boolean | null };


export const SendTestEmailMutationDocument = gql`
    mutation sendTestEmailMutation($appID: ID!, $smtpHost: String!, $smtpPort: Int!, $smtpUsername: String!, $smtpPassword: String!, $to: String!) {
  sendTestSMTPConfigurationEmail(
    input: {appID: $appID, smtpHost: $smtpHost, smtpPort: $smtpPort, smtpUsername: $smtpUsername, smtpPassword: $smtpPassword, to: $to}
  )
}
    `;
export type SendTestEmailMutationMutationFn = Apollo.MutationFunction<SendTestEmailMutationMutation, SendTestEmailMutationMutationVariables>;

/**
 * __useSendTestEmailMutationMutation__
 *
 * To run a mutation, you first call `useSendTestEmailMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSendTestEmailMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [sendTestEmailMutationMutation, { data, loading, error }] = useSendTestEmailMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      smtpHost: // value for 'smtpHost'
 *      smtpPort: // value for 'smtpPort'
 *      smtpUsername: // value for 'smtpUsername'
 *      smtpPassword: // value for 'smtpPassword'
 *      to: // value for 'to'
 *   },
 * });
 */
export function useSendTestEmailMutationMutation(baseOptions?: Apollo.MutationHookOptions<SendTestEmailMutationMutation, SendTestEmailMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SendTestEmailMutationMutation, SendTestEmailMutationMutationVariables>(SendTestEmailMutationDocument, options);
      }
export type SendTestEmailMutationMutationHookResult = ReturnType<typeof useSendTestEmailMutationMutation>;
export type SendTestEmailMutationMutationResult = Apollo.MutationResult<SendTestEmailMutationMutation>;
export type SendTestEmailMutationMutationOptions = Apollo.BaseMutationOptions<SendTestEmailMutationMutation, SendTestEmailMutationMutationVariables>;