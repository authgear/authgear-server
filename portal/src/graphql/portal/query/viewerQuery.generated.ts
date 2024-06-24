import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ViewerQueryQueryVariables = Types.Exact<{ [key: string]: never; }>;


export type ViewerQueryQuery = { __typename?: 'Query', viewer?: { __typename?: 'Viewer', id: string, email?: string | null, formattedName?: string | null, projectQuota?: number | null, projectOwnerCount: number, geoIPCountryCode?: string | null, isOnboardingSurveyCompleted?: boolean | null } | null };


export const ViewerQueryDocument = gql`
    query viewerQuery {
  viewer {
    id
    email
    formattedName
    projectQuota
    projectOwnerCount
    geoIPCountryCode
    isOnboardingSurveyCompleted
  }
}
    `;

/**
 * __useViewerQueryQuery__
 *
 * To run a query within a React component, call `useViewerQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useViewerQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useViewerQueryQuery({
 *   variables: {
 *   },
 * });
 */
export function useViewerQueryQuery(baseOptions?: Apollo.QueryHookOptions<ViewerQueryQuery, ViewerQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ViewerQueryQuery, ViewerQueryQueryVariables>(ViewerQueryDocument, options);
      }
export function useViewerQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ViewerQueryQuery, ViewerQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ViewerQueryQuery, ViewerQueryQueryVariables>(ViewerQueryDocument, options);
        }
export function useViewerQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ViewerQueryQuery, ViewerQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ViewerQueryQuery, ViewerQueryQueryVariables>(ViewerQueryDocument, options);
        }
export type ViewerQueryQueryHookResult = ReturnType<typeof useViewerQueryQuery>;
export type ViewerQueryLazyQueryHookResult = ReturnType<typeof useViewerQueryLazyQuery>;
export type ViewerQuerySuspenseQueryHookResult = ReturnType<typeof useViewerQuerySuspenseQuery>;
export type ViewerQueryQueryResult = Apollo.QueryResult<ViewerQueryQuery, ViewerQueryQueryVariables>;