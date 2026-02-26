import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppListQueryQueryVariables = Types.Exact<{ [key: string]: never; }>;


export type AppListQueryQuery = { __typename?: 'Query', appList?: Array<{ __typename?: 'AppListItem', appID: string, publicOrigin: string }> | null };


export const AppListQueryDocument = gql`
    query appListQuery {
  appList {
    appID
    publicOrigin
  }
}
    `;

/**
 * __useAppListQueryQuery__
 *
 * To run a query within a React component, call `useAppListQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAppListQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAppListQueryQuery({
 *   variables: {
 *   },
 * });
 */
export function useAppListQueryQuery(baseOptions?: Apollo.QueryHookOptions<AppListQueryQuery, AppListQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AppListQueryQuery, AppListQueryQueryVariables>(AppListQueryDocument, options);
      }
export function useAppListQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AppListQueryQuery, AppListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AppListQueryQuery, AppListQueryQueryVariables>(AppListQueryDocument, options);
        }
// @ts-ignore
export function useAppListQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AppListQueryQuery, AppListQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AppListQueryQuery, AppListQueryQueryVariables>;
export function useAppListQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AppListQueryQuery, AppListQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AppListQueryQuery | undefined, AppListQueryQueryVariables>;
export function useAppListQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AppListQueryQuery, AppListQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AppListQueryQuery, AppListQueryQueryVariables>(AppListQueryDocument, options);
        }
export type AppListQueryQueryHookResult = ReturnType<typeof useAppListQueryQuery>;
export type AppListQueryLazyQueryHookResult = ReturnType<typeof useAppListQueryLazyQuery>;
export type AppListQuerySuspenseQueryHookResult = ReturnType<typeof useAppListQuerySuspenseQuery>;
export type AppListQueryQueryResult = Apollo.QueryResult<AppListQueryQuery, AppListQueryQueryVariables>;