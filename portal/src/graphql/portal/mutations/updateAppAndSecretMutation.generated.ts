import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateAppAndSecretConfigMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  appConfig?: Types.InputMaybe<Types.Scalars['AppConfig']['input']>;
  appConfigChecksum?: Types.InputMaybe<Types.Scalars['String']['input']>;
  secretConfigUpdateInstructions?: Types.InputMaybe<Types.SecretConfigUpdateInstructionsInput>;
  secretConfigUpdateInstructionsChecksum?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type UpdateAppAndSecretConfigMutationMutation = { __typename?: 'Mutation', updateApp: { __typename?: 'UpdateAppPayload', app: { __typename?: 'App', id: string, effectiveAppConfig: any, rawAppConfig: any, rawAppConfigChecksum: any, secretConfigChecksum: string, samlIdpEntityID: string, secretConfig: { __typename?: 'SecretConfig', oauthSSOProviderClientSecrets?: Array<{ __typename?: 'OAuthSSOProviderClientSecret', alias: string, clientSecret?: string | null }> | null, webhookSecret?: { __typename?: 'WebhookSecret', secret?: string | null } | null, adminAPISecrets?: Array<{ __typename?: 'AdminAPISecret', keyID: string, createdAt?: any | null, publicKeyPEM: string, privateKeyPEM?: string | null }> | null, smtpSecret?: { __typename?: 'SMTPSecret', host: string, port: number, username: string, password?: string | null, sender?: string | null } | null, oauthClientSecrets?: Array<{ __typename?: 'oauthClientSecretItem', clientID: string, keys?: Array<{ __typename?: 'oauthClientSecretKey', keyID: string, createdAt?: any | null, key: string }> | null }> | null, botProtectionProviderSecret?: { __typename?: 'BotProtectionProviderSecret', type: string, secretKey?: string | null } | null, samlIdpSigningSecrets?: { __typename?: 'SAMLIdpSigningSecrets', certificates: Array<{ __typename?: 'SAMLIdpSigningCertificate', certificateFingerprint: string, certificatePEM: string, keyID: string }> } | null, samlSpSigningSecrets?: Array<{ __typename?: 'SAMLSpSigningSecrets', clientID: string, certificates: Array<{ __typename?: 'samlSpSigningCertificate', certificateFingerprint: string, certificatePEM: string }> }> | null }, effectiveSecretConfig: { __typename?: 'EffectiveSecretConfig', oauthSSOProviderDemoSecrets?: Array<{ __typename?: 'OAuthSSOProviderDemoSecretItem', type: string }> | null } } } };


export const UpdateAppAndSecretConfigMutationDocument = gql`
    mutation updateAppAndSecretConfigMutation($appID: ID!, $appConfig: AppConfig, $appConfigChecksum: String, $secretConfigUpdateInstructions: SecretConfigUpdateInstructionsInput, $secretConfigUpdateInstructionsChecksum: String) {
  updateApp(
    input: {appID: $appID, appConfig: $appConfig, appConfigChecksum: $appConfigChecksum, secretConfigUpdateInstructions: $secretConfigUpdateInstructions, secretConfigUpdateInstructionsChecksum: $secretConfigUpdateInstructionsChecksum}
  ) {
    app {
      id
      effectiveAppConfig
      rawAppConfig
      rawAppConfigChecksum
      secretConfig {
        oauthSSOProviderClientSecrets {
          alias
          clientSecret
        }
        webhookSecret {
          secret
        }
        adminAPISecrets {
          keyID
          createdAt
          publicKeyPEM
          privateKeyPEM
        }
        smtpSecret {
          host
          port
          username
          password
          sender
        }
        oauthClientSecrets {
          clientID
          keys {
            keyID
            createdAt
            key
          }
        }
        botProtectionProviderSecret {
          type
          secretKey
        }
        samlIdpSigningSecrets {
          certificates {
            certificateFingerprint
            certificatePEM
            keyID
          }
        }
        samlSpSigningSecrets {
          clientID
          certificates {
            certificateFingerprint
            certificatePEM
          }
        }
      }
      secretConfigChecksum
      effectiveSecretConfig {
        oauthSSOProviderDemoSecrets {
          type
        }
      }
      samlIdpEntityID
    }
  }
}
    `;
export type UpdateAppAndSecretConfigMutationMutationFn = Apollo.MutationFunction<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>;

/**
 * __useUpdateAppAndSecretConfigMutationMutation__
 *
 * To run a mutation, you first call `useUpdateAppAndSecretConfigMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateAppAndSecretConfigMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateAppAndSecretConfigMutationMutation, { data, loading, error }] = useUpdateAppAndSecretConfigMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      appConfig: // value for 'appConfig'
 *      appConfigChecksum: // value for 'appConfigChecksum'
 *      secretConfigUpdateInstructions: // value for 'secretConfigUpdateInstructions'
 *      secretConfigUpdateInstructionsChecksum: // value for 'secretConfigUpdateInstructionsChecksum'
 *   },
 * });
 */
export function useUpdateAppAndSecretConfigMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>(UpdateAppAndSecretConfigMutationDocument, options);
      }
export type UpdateAppAndSecretConfigMutationMutationHookResult = ReturnType<typeof useUpdateAppAndSecretConfigMutationMutation>;
export type UpdateAppAndSecretConfigMutationMutationResult = Apollo.MutationResult<UpdateAppAndSecretConfigMutationMutation>;
export type UpdateAppAndSecretConfigMutationMutationOptions = Apollo.BaseMutationOptions<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>;