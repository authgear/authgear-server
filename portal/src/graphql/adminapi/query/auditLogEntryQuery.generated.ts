import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AuditLogEntryFragment = { __typename?: 'AuditLog', id: string, createdAt: any, activityType: Types.AuditLogActivityType, ipAddress?: string | null, userAgent?: string | null, clientID?: string | null, data?: any | null, user?: { __typename?: 'User', id: string } | null };

export type AuditLogEntryQueryQueryVariables = Types.Exact<{
  logID: Types.Scalars['ID']['input'];
}>;


export type AuditLogEntryQueryQuery = { __typename?: 'Query', node?: { __typename: 'AuditLog', id: string, createdAt: any, activityType: Types.AuditLogActivityType, ipAddress?: string | null, userAgent?: string | null, clientID?: string | null, data?: any | null, user?: { __typename?: 'User', id: string } | null } | { __typename: 'Authenticator' } | { __typename: 'Authorization' } | { __typename: 'Group' } | { __typename: 'Identity' } | { __typename: 'Resource' } | { __typename: 'Role' } | { __typename: 'Scope' } | { __typename: 'Session' } | { __typename: 'User' } | null };

export const AuditLogEntryFragmentDoc = gql`
    fragment AuditLogEntry on AuditLog {
  id
  createdAt
  activityType
  user {
    id
  }
  ipAddress
  userAgent
  clientID
  data
}
    `;
export const AuditLogEntryQueryDocument = gql`
    query AuditLogEntryQuery($logID: ID!) {
  node(id: $logID) {
    __typename
    ...AuditLogEntry
  }
}
    ${AuditLogEntryFragmentDoc}`;

/**
 * __useAuditLogEntryQueryQuery__
 *
 * To run a query within a React component, call `useAuditLogEntryQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAuditLogEntryQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAuditLogEntryQueryQuery({
 *   variables: {
 *      logID: // value for 'logID'
 *   },
 * });
 */
export function useAuditLogEntryQueryQuery(baseOptions: Apollo.QueryHookOptions<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables> & ({ variables: AuditLogEntryQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>(AuditLogEntryQueryDocument, options);
      }
export function useAuditLogEntryQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>(AuditLogEntryQueryDocument, options);
        }
// @ts-ignore
export function useAuditLogEntryQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>;
export function useAuditLogEntryQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AuditLogEntryQueryQuery | undefined, AuditLogEntryQueryQueryVariables>;
export function useAuditLogEntryQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>(AuditLogEntryQueryDocument, options);
        }
export type AuditLogEntryQueryQueryHookResult = ReturnType<typeof useAuditLogEntryQueryQuery>;
export type AuditLogEntryQueryLazyQueryHookResult = ReturnType<typeof useAuditLogEntryQueryLazyQuery>;
export type AuditLogEntryQuerySuspenseQueryHookResult = ReturnType<typeof useAuditLogEntryQuerySuspenseQuery>;
export type AuditLogEntryQueryQueryResult = Apollo.QueryResult<AuditLogEntryQueryQuery, AuditLogEntryQueryQueryVariables>;