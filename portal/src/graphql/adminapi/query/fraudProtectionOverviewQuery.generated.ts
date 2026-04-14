import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type FraudProtectionOverviewQueryQueryVariables = Types.Exact<{
  rangeFrom?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  rangeTo?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
}>;


export type FraudProtectionOverviewQueryQuery = { __typename?: 'Query', fraudProtectionOverview: { __typename?: 'FraudProtectionOverview', sendSMS: { __typename?: 'FraudProtectionOverviewSendSMS', total: number, blocked: number, flagged: number, topSourceIPs: Array<{ __typename?: 'FraudProtectionOverviewTopSourceIP', ipAddress: string, total: number, blocked: number, flagged: number }> } } };


export const FraudProtectionOverviewQueryDocument = gql`
    query fraudProtectionOverviewQuery($rangeFrom: DateTime, $rangeTo: DateTime) {
  fraudProtectionOverview(rangeFrom: $rangeFrom, rangeTo: $rangeTo) {
    sendSMS {
      total
      blocked
      flagged
      topSourceIPs {
        ipAddress
        total
        blocked
        flagged
      }
    }
  }
}
    `;

/**
 * __useFraudProtectionOverviewQueryQuery__
 *
 * To run a query within a React component, call `useFraudProtectionOverviewQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useFraudProtectionOverviewQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFraudProtectionOverviewQueryQuery({
 *   variables: {
 *      rangeFrom: // value for 'rangeFrom'
 *      rangeTo: // value for 'rangeTo'
 *   },
 * });
 */
export function useFraudProtectionOverviewQueryQuery(baseOptions?: Apollo.QueryHookOptions<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>(FraudProtectionOverviewQueryDocument, options);
      }
export function useFraudProtectionOverviewQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>(FraudProtectionOverviewQueryDocument, options);
        }
// @ts-ignore
export function useFraudProtectionOverviewQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>): Apollo.UseSuspenseQueryResult<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>;
export function useFraudProtectionOverviewQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>): Apollo.UseSuspenseQueryResult<FraudProtectionOverviewQueryQuery | undefined, FraudProtectionOverviewQueryQueryVariables>;
export function useFraudProtectionOverviewQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>(FraudProtectionOverviewQueryDocument, options);
        }
export type FraudProtectionOverviewQueryQueryHookResult = ReturnType<typeof useFraudProtectionOverviewQueryQuery>;
export type FraudProtectionOverviewQueryLazyQueryHookResult = ReturnType<typeof useFraudProtectionOverviewQueryLazyQuery>;
export type FraudProtectionOverviewQuerySuspenseQueryHookResult = ReturnType<typeof useFraudProtectionOverviewQuerySuspenseQuery>;
export type FraudProtectionOverviewQueryQueryResult = Apollo.QueryResult<FraudProtectionOverviewQueryQuery, FraudProtectionOverviewQueryQueryVariables>;