import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppAndSecretConfigFragment = { __typename?: 'App', id: string, effectiveAppConfig: any, rawAppConfig: any, rawAppConfigChecksum: any, secretConfigChecksum: string, samlIdpEntityID: string, secretConfig: { __typename?: 'SecretConfig', oauthSSOProviderClientSecrets?: Array<{ __typename?: 'OAuthSSOProviderClientSecret', alias: string, clientSecret?: string | null }> | null, webhookSecret?: { __typename?: 'WebhookSecret', secret?: string | null } | null, adminAPISecrets?: Array<{ __typename?: 'AdminAPISecret', keyID: string, createdAt?: any | null, publicKeyPEM: string, privateKeyPEM?: string | null }> | null, smtpSecret?: { __typename?: 'SMTPSecret', host: string, port: number, username: string, password?: string | null, sender?: string | null } | null, oauthClientSecrets?: Array<{ __typename?: 'oauthClientSecretItem', clientID: string, keys?: Array<{ __typename?: 'oauthClientSecretKey', keyID: string, createdAt?: any | null, key: string }> | null }> | null, botProtectionProviderSecret?: { __typename?: 'BotProtectionProviderSecret', type: string, secretKey?: string | null } | null, samlIdpSigningSecrets?: { __typename?: 'SAMLIdpSigningSecrets', certificates: Array<{ __typename?: 'SAMLIdpSigningCertificate', certificateFingerprint: string, certificatePEM: string, keyID: string }> } | null, samlSpSigningSecrets?: Array<{ __typename?: 'SAMLSpSigningSecrets', clientID: string, certificates: Array<{ __typename?: 'samlSpSigningCertificate', certificateFingerprint: string, certificatePEM: string }> }> | null, smsProviderSecrets?: { __typename?: 'SMSProviderSecrets', customSMSProviderCredentials?: { __typename?: 'SMSProviderCustomSMSProviderSecrets', timeout?: number | null, url: string } | null, twilioCredentials?: { __typename?: 'SMSProviderTwilioCredentials', credentialType: Types.TwilioCredentialType, accountSID: string, authToken?: string | null, apiKeySID?: string | null, apiKeySecret?: string | null, messagingServiceSID?: string | null, from?: string | null } | null } | null }, effectiveSecretConfig: { __typename?: 'EffectiveSecretConfig', oauthSSOProviderDemoSecrets?: Array<{ __typename?: 'OAuthSSOProviderDemoSecretItem', type: string }> | null }, viewer: { __typename?: 'Collaborator', id: string, role: Types.CollaboratorRole, createdAt: any, user: { __typename?: 'User', id: string, email?: string | null } } };

export type AppAndSecretConfigQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
  token?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type AppAndSecretConfigQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveAppConfig: any, rawAppConfig: any, rawAppConfigChecksum: any, secretConfigChecksum: string, samlIdpEntityID: string, secretConfig: { __typename?: 'SecretConfig', oauthSSOProviderClientSecrets?: Array<{ __typename?: 'OAuthSSOProviderClientSecret', alias: string, clientSecret?: string | null }> | null, webhookSecret?: { __typename?: 'WebhookSecret', secret?: string | null } | null, adminAPISecrets?: Array<{ __typename?: 'AdminAPISecret', keyID: string, createdAt?: any | null, publicKeyPEM: string, privateKeyPEM?: string | null }> | null, smtpSecret?: { __typename?: 'SMTPSecret', host: string, port: number, username: string, password?: string | null, sender?: string | null } | null, oauthClientSecrets?: Array<{ __typename?: 'oauthClientSecretItem', clientID: string, keys?: Array<{ __typename?: 'oauthClientSecretKey', keyID: string, createdAt?: any | null, key: string }> | null }> | null, botProtectionProviderSecret?: { __typename?: 'BotProtectionProviderSecret', type: string, secretKey?: string | null } | null, samlIdpSigningSecrets?: { __typename?: 'SAMLIdpSigningSecrets', certificates: Array<{ __typename?: 'SAMLIdpSigningCertificate', certificateFingerprint: string, certificatePEM: string, keyID: string }> } | null, samlSpSigningSecrets?: Array<{ __typename?: 'SAMLSpSigningSecrets', clientID: string, certificates: Array<{ __typename?: 'samlSpSigningCertificate', certificateFingerprint: string, certificatePEM: string }> }> | null, smsProviderSecrets?: { __typename?: 'SMSProviderSecrets', customSMSProviderCredentials?: { __typename?: 'SMSProviderCustomSMSProviderSecrets', timeout?: number | null, url: string } | null, twilioCredentials?: { __typename?: 'SMSProviderTwilioCredentials', credentialType: Types.TwilioCredentialType, accountSID: string, authToken?: string | null, apiKeySID?: string | null, apiKeySecret?: string | null, messagingServiceSID?: string | null, from?: string | null } | null } | null }, effectiveSecretConfig: { __typename?: 'EffectiveSecretConfig', oauthSSOProviderDemoSecrets?: Array<{ __typename?: 'OAuthSSOProviderDemoSecretItem', type: string }> | null }, viewer: { __typename?: 'Collaborator', id: string, role: Types.CollaboratorRole, createdAt: any, user: { __typename?: 'User', id: string, email?: string | null } } } | { __typename: 'User' } | { __typename: 'Viewer' } | null };

export const AppAndSecretConfigFragmentDoc = gql`
    fragment AppAndSecretConfig on App {
  id
  effectiveAppConfig
  rawAppConfig
  rawAppConfigChecksum
  secretConfig(token: $token) {
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
    smsProviderSecrets {
      customSMSProviderCredentials {
        timeout
        url
      }
      twilioCredentials {
        credentialType
        accountSID
        authToken
        apiKeySID
        apiKeySecret
        messagingServiceSID
        from
      }
    }
  }
  secretConfigChecksum
  effectiveSecretConfig {
    oauthSSOProviderDemoSecrets {
      type
    }
  }
  viewer {
    id
    role
    createdAt
    user {
      id
      email
    }
  }
  samlIdpEntityID
}
    `;
export const AppAndSecretConfigQueryDocument = gql`
    query appAndSecretConfigQuery($id: ID!, $token: String) {
  node(id: $id) {
    __typename
    ...AppAndSecretConfig
  }
}
    ${AppAndSecretConfigFragmentDoc}`;

/**
 * __useAppAndSecretConfigQueryQuery__
 *
 * To run a query within a React component, call `useAppAndSecretConfigQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAppAndSecretConfigQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAppAndSecretConfigQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *      token: // value for 'token'
 *   },
 * });
 */
export function useAppAndSecretConfigQueryQuery(baseOptions: Apollo.QueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables> & ({ variables: AppAndSecretConfigQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>(AppAndSecretConfigQueryDocument, options);
      }
export function useAppAndSecretConfigQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>(AppAndSecretConfigQueryDocument, options);
        }
// @ts-ignore
export function useAppAndSecretConfigQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>;
export function useAppAndSecretConfigQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AppAndSecretConfigQueryQuery | undefined, AppAndSecretConfigQueryQueryVariables>;
export function useAppAndSecretConfigQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>(AppAndSecretConfigQueryDocument, options);
        }
export type AppAndSecretConfigQueryQueryHookResult = ReturnType<typeof useAppAndSecretConfigQueryQuery>;
export type AppAndSecretConfigQueryLazyQueryHookResult = ReturnType<typeof useAppAndSecretConfigQueryLazyQuery>;
export type AppAndSecretConfigQuerySuspenseQueryHookResult = ReturnType<typeof useAppAndSecretConfigQuerySuspenseQuery>;
export type AppAndSecretConfigQueryQueryResult = Apollo.QueryResult<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>;