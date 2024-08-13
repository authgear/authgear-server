import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UserQueryNodeFragment = { __typename?: 'User', id: string, standardAttributes: any, customAttributes: any, web3: any, formattedName?: string | null, endUserAccountID?: string | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, lastLoginAt?: any | null, createdAt: any, updatedAt: any, mfaGracePeriodEndAt?: any | null, roles?: { __typename?: 'RoleConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'RoleEdge', cursor: string, node?: { __typename?: 'Role', createdAt: any, description?: string | null, id: string, key: string, name?: string | null, updatedAt: any } | null } | null> | null } | null, groups?: { __typename?: 'GroupConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'GroupEdge', cursor: string, node?: { __typename?: 'Group', createdAt: any, description?: string | null, id: string, key: string, name?: string | null, updatedAt: any, roles?: { __typename?: 'RoleConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'RoleEdge', node?: { __typename?: 'Role', createdAt: any, description?: string | null, id: string, key: string, name?: string | null, updatedAt: any } | null } | null> | null } | null } | null } | null> | null } | null, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string, type: Types.AuthenticatorType, kind: Types.AuthenticatorKind, isDefault: boolean, claims: any, createdAt: any, updatedAt: any, expireAfter?: any | null } | null } | null> | null } | null, identities?: { __typename?: 'IdentityConnection', edges?: Array<{ __typename?: 'IdentityEdge', node?: { __typename?: 'Identity', id: string, type: Types.IdentityType, claims: any, createdAt: any, updatedAt: any } | null } | null> | null } | null, verifiedClaims: Array<{ __typename?: 'Claim', name: string, value: string }>, sessions?: { __typename?: 'SessionConnection', edges?: Array<{ __typename?: 'SessionEdge', node?: { __typename?: 'Session', id: string, type: Types.SessionType, clientID?: string | null, lastAccessedAt: any, lastAccessedByIP: string, displayName: string, userAgent?: string | null } | null } | null> | null } | null, authorizations?: { __typename?: 'AuthorizationConnection', edges?: Array<{ __typename?: 'AuthorizationEdge', node?: { __typename?: 'Authorization', id: string, clientID: string, scopes: Array<string>, createdAt: any } | null } | null> | null } | null };

export type AuthenticatorFragmentFragment = { __typename?: 'Authenticator', id: string, type: Types.AuthenticatorType, kind: Types.AuthenticatorKind, isDefault: boolean, claims: any, createdAt: any, updatedAt: any, expireAfter?: any | null };

export type UserQueryQueryVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type UserQueryQuery = { __typename?: 'Query', node?: { __typename: 'AuditLog' } | { __typename: 'Authenticator' } | { __typename: 'Authorization' } | { __typename: 'Group' } | { __typename: 'Identity' } | { __typename: 'Role' } | { __typename: 'Session' } | { __typename: 'User', id: string, standardAttributes: any, customAttributes: any, web3: any, formattedName?: string | null, endUserAccountID?: string | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, lastLoginAt?: any | null, createdAt: any, updatedAt: any, mfaGracePeriodEndAt?: any | null, roles?: { __typename?: 'RoleConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'RoleEdge', cursor: string, node?: { __typename?: 'Role', createdAt: any, description?: string | null, id: string, key: string, name?: string | null, updatedAt: any } | null } | null> | null } | null, groups?: { __typename?: 'GroupConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'GroupEdge', cursor: string, node?: { __typename?: 'Group', createdAt: any, description?: string | null, id: string, key: string, name?: string | null, updatedAt: any, roles?: { __typename?: 'RoleConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'RoleEdge', node?: { __typename?: 'Role', createdAt: any, description?: string | null, id: string, key: string, name?: string | null, updatedAt: any } | null } | null> | null } | null } | null } | null> | null } | null, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string, type: Types.AuthenticatorType, kind: Types.AuthenticatorKind, isDefault: boolean, claims: any, createdAt: any, updatedAt: any, expireAfter?: any | null } | null } | null> | null } | null, identities?: { __typename?: 'IdentityConnection', edges?: Array<{ __typename?: 'IdentityEdge', node?: { __typename?: 'Identity', id: string, type: Types.IdentityType, claims: any, createdAt: any, updatedAt: any } | null } | null> | null } | null, verifiedClaims: Array<{ __typename?: 'Claim', name: string, value: string }>, sessions?: { __typename?: 'SessionConnection', edges?: Array<{ __typename?: 'SessionEdge', node?: { __typename?: 'Session', id: string, type: Types.SessionType, clientID?: string | null, lastAccessedAt: any, lastAccessedByIP: string, displayName: string, userAgent?: string | null } | null } | null> | null } | null, authorizations?: { __typename?: 'AuthorizationConnection', edges?: Array<{ __typename?: 'AuthorizationEdge', node?: { __typename?: 'Authorization', id: string, clientID: string, scopes: Array<string>, createdAt: any } | null } | null> | null } | null } | null };

export const AuthenticatorFragmentFragmentDoc = gql`
    fragment AuthenticatorFragment on Authenticator {
  id
  type
  kind
  isDefault
  claims
  createdAt
  updatedAt
  expireAfter
}
    `;
export const UserQueryNodeFragmentDoc = gql`
    fragment UserQueryNode on User {
  id
  roles {
    totalCount
    edges {
      cursor
      node {
        createdAt
        description
        id
        key
        name
        updatedAt
      }
    }
  }
  groups {
    totalCount
    edges {
      cursor
      node {
        createdAt
        description
        id
        key
        roles {
          edges {
            node {
              createdAt
              description
              id
              key
              name
              updatedAt
            }
          }
        }
        name
        roles {
          totalCount
          edges {
            node {
              createdAt
              description
              id
              key
              name
              updatedAt
            }
          }
        }
        updatedAt
      }
    }
  }
  authenticators {
    edges {
      node {
        ...AuthenticatorFragment
      }
    }
  }
  identities {
    edges {
      node {
        id
        type
        claims
        createdAt
        updatedAt
      }
    }
  }
  verifiedClaims {
    name
    value
  }
  standardAttributes
  customAttributes
  web3
  sessions {
    edges {
      node {
        id
        type
        clientID
        lastAccessedAt
        lastAccessedByIP
        displayName
        userAgent
      }
    }
  }
  authorizations {
    edges {
      node {
        id
        clientID
        scopes
        createdAt
      }
    }
  }
  formattedName
  endUserAccountID
  isAnonymous
  isDisabled
  disableReason
  isDeactivated
  deleteAt
  isAnonymized
  anonymizeAt
  lastLoginAt
  createdAt
  updatedAt
  mfaGracePeriodEndAt
}
    ${AuthenticatorFragmentFragmentDoc}`;
export const UserQueryDocument = gql`
    query userQuery($userID: ID!) {
  node(id: $userID) {
    __typename
    ...UserQueryNode
  }
}
    ${UserQueryNodeFragmentDoc}`;

/**
 * __useUserQueryQuery__
 *
 * To run a query within a React component, call `useUserQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useUserQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUserQueryQuery({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useUserQueryQuery(baseOptions: Apollo.QueryHookOptions<UserQueryQuery, UserQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<UserQueryQuery, UserQueryQueryVariables>(UserQueryDocument, options);
      }
export function useUserQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<UserQueryQuery, UserQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<UserQueryQuery, UserQueryQueryVariables>(UserQueryDocument, options);
        }
export function useUserQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<UserQueryQuery, UserQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<UserQueryQuery, UserQueryQueryVariables>(UserQueryDocument, options);
        }
export type UserQueryQueryHookResult = ReturnType<typeof useUserQueryQuery>;
export type UserQueryLazyQueryHookResult = ReturnType<typeof useUserQueryLazyQuery>;
export type UserQuerySuspenseQueryHookResult = ReturnType<typeof useUserQuerySuspenseQuery>;
export type UserQueryQueryResult = Apollo.QueryResult<UserQueryQuery, UserQueryQueryVariables>;