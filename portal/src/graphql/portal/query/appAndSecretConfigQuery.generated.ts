import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppAndSecretConfigFragment = { __typename?: 'App', id: string, effectiveAppConfig: any, rawAppConfig: any, secretConfig: { __typename?: 'SecretConfig', oauthClientSecrets?: Array<{ __typename?: 'OAuthClientSecret', alias: string, clientSecret: string }> | null, webhookSecret?: { __typename?: 'WebhookSecret', secret?: string | null } | null, adminAPISecrets?: Array<{ __typename?: 'AdminAPISecret', keyID: string, createdAt?: any | null, publicKeyPEM: string, privateKeyPEM?: string | null }> | null, smtpSecret?: { __typename?: 'SMTPSecret', host: string, port: number, username: string, password?: string | null } | null } };

export type AppAndSecretConfigQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID'];
}>;


export type AppAndSecretConfigQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveAppConfig: any, rawAppConfig: any, secretConfig: { __typename?: 'SecretConfig', oauthClientSecrets?: Array<{ __typename?: 'OAuthClientSecret', alias: string, clientSecret: string }> | null, webhookSecret?: { __typename?: 'WebhookSecret', secret?: string | null } | null, adminAPISecrets?: Array<{ __typename?: 'AdminAPISecret', keyID: string, createdAt?: any | null, publicKeyPEM: string, privateKeyPEM?: string | null }> | null, smtpSecret?: { __typename?: 'SMTPSecret', host: string, port: number, username: string, password?: string | null } | null } } | { __typename: 'User' } | null };


export const AppAndSecretConfigQueryDocument = gql`
    query appAndSecretConfigQuery($id: ID!) {
  node(id: $id) {
    __typename
    ... on App {
      id
      effectiveAppConfig
      rawAppConfig
      secretConfig {
        oauthClientSecrets {
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
        }
      }
    }
  }
}
    `;

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
 *   },
 * });
 */
export function useAppAndSecretConfigQueryQuery(baseOptions: Apollo.QueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>(AppAndSecretConfigQueryDocument, options);
      }
export function useAppAndSecretConfigQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>(AppAndSecretConfigQueryDocument, options);
        }
export type AppAndSecretConfigQueryQueryHookResult = ReturnType<typeof useAppAndSecretConfigQueryQuery>;
export type AppAndSecretConfigQueryLazyQueryHookResult = ReturnType<typeof useAppAndSecretConfigQueryLazyQuery>;
export type AppAndSecretConfigQueryQueryResult = Apollo.QueryResult<AppAndSecretConfigQueryQuery, AppAndSecretConfigQueryQueryVariables>;