import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ViewerQueryQueryVariables = Types.Exact<{ [key: string]: never; }>;


export type ViewerQueryQuery = { __typename?: 'Query', viewer?: { __typename?: 'User', id: string, email?: string | null } | null };


export const ViewerQueryDocument = gql`
    query viewerQuery {
  viewer {
    id
    email
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
export type ViewerQueryQueryHookResult = ReturnType<typeof useViewerQueryQuery>;
export type ViewerQueryLazyQueryHookResult = ReturnType<typeof useViewerQueryLazyQuery>;
export type ViewerQueryQueryResult = Apollo.QueryResult<ViewerQueryQuery, ViewerQueryQueryVariables>;