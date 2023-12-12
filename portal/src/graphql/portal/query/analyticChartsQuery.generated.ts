import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AnalyticChartsQueryQueryVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  periodical: Types.Periodical;
  rangeFrom: Types.Scalars['Date']['input'];
  rangeTo: Types.Scalars['Date']['input'];
}>;


export type AnalyticChartsQueryQuery = { __typename?: 'Query', activeUserChart?: { __typename?: 'Chart', dataset: Array<{ __typename?: 'DataPoint', label: string, data: number } | null> } | null, totalUserCountChart?: { __typename?: 'Chart', dataset: Array<{ __typename?: 'DataPoint', label: string, data: number } | null> } | null, signupConversionRate?: { __typename?: 'SignupConversionRate', totalSignup: number, totalSignupUniquePageView: number } | null, signupByMethodsChart?: { __typename?: 'Chart', dataset: Array<{ __typename?: 'DataPoint', label: string, data: number } | null> } | null };


export const AnalyticChartsQueryDocument = gql`
    query analyticChartsQuery($appID: ID!, $periodical: Periodical!, $rangeFrom: Date!, $rangeTo: Date!) {
  activeUserChart(
    appID: $appID
    periodical: $periodical
    rangeFrom: $rangeFrom
    rangeTo: $rangeTo
  ) {
    dataset {
      label
      data
    }
  }
  totalUserCountChart(appID: $appID, rangeFrom: $rangeFrom, rangeTo: $rangeTo) {
    dataset {
      label
      data
    }
  }
  signupConversionRate(appID: $appID, rangeFrom: $rangeFrom, rangeTo: $rangeTo) {
    totalSignup
    totalSignupUniquePageView
  }
  signupByMethodsChart(appID: $appID, rangeFrom: $rangeFrom, rangeTo: $rangeTo) {
    dataset {
      label
      data
    }
  }
}
    `;

/**
 * __useAnalyticChartsQueryQuery__
 *
 * To run a query within a React component, call `useAnalyticChartsQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAnalyticChartsQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAnalyticChartsQueryQuery({
 *   variables: {
 *      appID: // value for 'appID'
 *      periodical: // value for 'periodical'
 *      rangeFrom: // value for 'rangeFrom'
 *      rangeTo: // value for 'rangeTo'
 *   },
 * });
 */
export function useAnalyticChartsQueryQuery(baseOptions: Apollo.QueryHookOptions<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>(AnalyticChartsQueryDocument, options);
      }
export function useAnalyticChartsQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>(AnalyticChartsQueryDocument, options);
        }
export function useAnalyticChartsQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>(AnalyticChartsQueryDocument, options);
        }
export type AnalyticChartsQueryQueryHookResult = ReturnType<typeof useAnalyticChartsQueryQuery>;
export type AnalyticChartsQueryLazyQueryHookResult = ReturnType<typeof useAnalyticChartsQueryLazyQuery>;
export type AnalyticChartsQuerySuspenseQueryHookResult = ReturnType<typeof useAnalyticChartsQuerySuspenseQuery>;
export type AnalyticChartsQueryQueryResult = Apollo.QueryResult<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>;