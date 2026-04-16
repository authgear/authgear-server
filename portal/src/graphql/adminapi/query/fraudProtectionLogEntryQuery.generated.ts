import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type FraudProtectionLogEntryFragment = { __typename?: 'FraudProtectionDecisionRecord', id: string, createdAt: any, decision: Types.FraudProtectionDecision, action: Types.FraudProtectionAction, triggeredWarnings: Array<Types.FraudProtectionWarningType>, userAgent?: string | null, ipAddress?: string | null, geoLocationCode?: string | null, data?: any | null, actionDetail: { __typename: 'FraudProtectionDecisionSendSMSActionDetail', recipient: string, type: string, phoneNumberCountryCode?: string | null } };

export type FraudProtectionLogEntryQueryQueryVariables = Types.Exact<{
  logID: Types.Scalars['ID']['input'];
}>;


export type FraudProtectionLogEntryQueryQuery = { __typename?: 'Query', node?: { __typename: 'AuditLog' } | { __typename: 'Authenticator' } | { __typename: 'Authorization' } | { __typename: 'FraudProtectionDecisionRecord', id: string, createdAt: any, decision: Types.FraudProtectionDecision, action: Types.FraudProtectionAction, triggeredWarnings: Array<Types.FraudProtectionWarningType>, userAgent?: string | null, ipAddress?: string | null, geoLocationCode?: string | null, data?: any | null, actionDetail: { __typename: 'FraudProtectionDecisionSendSMSActionDetail', recipient: string, type: string, phoneNumberCountryCode?: string | null } } | { __typename: 'Group' } | { __typename: 'Identity' } | { __typename: 'Resource' } | { __typename: 'Role' } | { __typename: 'Scope' } | { __typename: 'Session' } | { __typename: 'User' } | null };

export const FraudProtectionLogEntryFragmentDoc = gql`
    fragment FraudProtectionLogEntry on FraudProtectionDecisionRecord {
  id
  createdAt
  decision
  action
  actionDetail {
    __typename
    ... on FraudProtectionDecisionSendSMSActionDetail {
      recipient
      type
      phoneNumberCountryCode
    }
  }
  triggeredWarnings
  userAgent
  ipAddress
  geoLocationCode
  data
}
    `;
export const FraudProtectionLogEntryQueryDocument = gql`
    query FraudProtectionLogEntryQuery($logID: ID!) {
  node(id: $logID) {
    __typename
    ...FraudProtectionLogEntry
  }
}
    ${FraudProtectionLogEntryFragmentDoc}`;

/**
 * __useFraudProtectionLogEntryQueryQuery__
 *
 * To run a query within a React component, call `useFraudProtectionLogEntryQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useFraudProtectionLogEntryQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFraudProtectionLogEntryQueryQuery({
 *   variables: {
 *      logID: // value for 'logID'
 *   },
 * });
 */
export function useFraudProtectionLogEntryQueryQuery(baseOptions: Apollo.QueryHookOptions<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables> & ({ variables: FraudProtectionLogEntryQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>(FraudProtectionLogEntryQueryDocument, options);
      }
export function useFraudProtectionLogEntryQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>(FraudProtectionLogEntryQueryDocument, options);
        }
// @ts-ignore
export function useFraudProtectionLogEntryQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>): Apollo.UseSuspenseQueryResult<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>;
export function useFraudProtectionLogEntryQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>): Apollo.UseSuspenseQueryResult<FraudProtectionLogEntryQueryQuery | undefined, FraudProtectionLogEntryQueryQueryVariables>;
export function useFraudProtectionLogEntryQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>(FraudProtectionLogEntryQueryDocument, options);
        }
export type FraudProtectionLogEntryQueryQueryHookResult = ReturnType<typeof useFraudProtectionLogEntryQueryQuery>;
export type FraudProtectionLogEntryQueryLazyQueryHookResult = ReturnType<typeof useFraudProtectionLogEntryQueryLazyQuery>;
export type FraudProtectionLogEntryQuerySuspenseQueryHookResult = ReturnType<typeof useFraudProtectionLogEntryQuerySuspenseQuery>;
export type FraudProtectionLogEntryQueryQueryResult = Apollo.QueryResult<FraudProtectionLogEntryQueryQuery, FraudProtectionLogEntryQueryQueryVariables>;