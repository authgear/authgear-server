import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type FraudProtectionLogsQueryQueryVariables = Types.Exact<{
  pageSize: Types.Scalars['Int']['input'];
  cursor?: Types.InputMaybe<Types.Scalars['String']['input']>;
  rangeFrom?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  rangeTo?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  sortDirection?: Types.InputMaybe<Types.SortDirection>;
  verdicts?: Types.InputMaybe<Array<Types.FraudProtectionDecision> | Types.FraudProtectionDecision>;
  reasonCodes?: Types.InputMaybe<Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input']>;
  maximumWarningCount?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  minimumWarningCount?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  search?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type FraudProtectionLogsQueryQuery = { __typename?: 'Query', fraudProtectionLogs?: { __typename?: 'FraudProtectionDecisionRecordConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'FraudProtectionDecisionRecordEdge', node?: { __typename?: 'FraudProtectionDecisionRecord', id: string, createdAt: any, decision: Types.FraudProtectionDecision, action: Types.FraudProtectionAction, triggeredWarnings: Array<Types.FraudProtectionWarningType>, userAgent?: string | null, ipAddress?: string | null, geoLocationCode?: string | null, actionDetail: { __typename: 'FraudProtectionDecisionSendSMSActionDetail', recipient: string, type: string, phoneNumberCountryCode?: string | null } } | null } | null> | null } | null };


export const FraudProtectionLogsQueryDocument = gql`
    query FraudProtectionLogsQuery($pageSize: Int!, $cursor: String, $rangeFrom: DateTime, $rangeTo: DateTime, $sortDirection: SortDirection, $verdicts: [FraudProtectionDecision!], $reasonCodes: [String!], $maximumWarningCount: Int, $minimumWarningCount: Int, $search: String) {
  fraudProtectionLogs(
    first: $pageSize
    after: $cursor
    rangeFrom: $rangeFrom
    rangeTo: $rangeTo
    sortDirection: $sortDirection
    verdicts: $verdicts
    reasonCodes: $reasonCodes
    maximumWarningCount: $maximumWarningCount
    minimumWarningCount: $minimumWarningCount
    search: $search
  ) {
    edges {
      node {
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
      }
    }
    totalCount
  }
}
    `;

/**
 * __useFraudProtectionLogsQueryQuery__
 *
 * To run a query within a React component, call `useFraudProtectionLogsQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useFraudProtectionLogsQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFraudProtectionLogsQueryQuery({
 *   variables: {
 *      pageSize: // value for 'pageSize'
 *      cursor: // value for 'cursor'
 *      rangeFrom: // value for 'rangeFrom'
 *      rangeTo: // value for 'rangeTo'
 *      sortDirection: // value for 'sortDirection'
 *      verdicts: // value for 'verdicts'
 *      reasonCodes: // value for 'reasonCodes'
 *      maximumWarningCount: // value for 'maximumWarningCount'
 *      minimumWarningCount: // value for 'minimumWarningCount'
 *      search: // value for 'search'
 *   },
 * });
 */
export function useFraudProtectionLogsQueryQuery(baseOptions: Apollo.QueryHookOptions<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables> & ({ variables: FraudProtectionLogsQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>(FraudProtectionLogsQueryDocument, options);
      }
export function useFraudProtectionLogsQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>(FraudProtectionLogsQueryDocument, options);
        }
// @ts-ignore
export function useFraudProtectionLogsQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>): Apollo.UseSuspenseQueryResult<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>;
export function useFraudProtectionLogsQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>): Apollo.UseSuspenseQueryResult<FraudProtectionLogsQueryQuery | undefined, FraudProtectionLogsQueryQueryVariables>;
export function useFraudProtectionLogsQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>(FraudProtectionLogsQueryDocument, options);
        }
export type FraudProtectionLogsQueryQueryHookResult = ReturnType<typeof useFraudProtectionLogsQueryQuery>;
export type FraudProtectionLogsQueryLazyQueryHookResult = ReturnType<typeof useFraudProtectionLogsQueryLazyQuery>;
export type FraudProtectionLogsQuerySuspenseQueryHookResult = ReturnType<typeof useFraudProtectionLogsQuerySuspenseQuery>;
export type FraudProtectionLogsQueryQueryResult = Apollo.QueryResult<FraudProtectionLogsQueryQuery, FraudProtectionLogsQueryQueryVariables>;