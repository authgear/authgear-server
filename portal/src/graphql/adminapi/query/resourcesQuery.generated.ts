import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ResourcesQueryQueryVariables = Types.Exact<{
  first?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  after?: Types.InputMaybe<Types.Scalars['String']['input']>;
  clientID?: Types.InputMaybe<Types.Scalars['String']['input']>;
  searchKeyword?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type ResourcesQueryQuery = { __typename?: 'Query', resources?: { __typename?: 'ResourceConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'ResourceEdge', cursor: string, node?: { __typename?: 'Resource', id: string, name?: string | null, clientIDs: Array<string>, resourceURI: string, createdAt: any, updatedAt: any } | null } | null> | null, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null } } | null };


export const ResourcesQueryDocument = gql`
    query resourcesQuery($first: Int, $after: String, $clientID: String, $searchKeyword: String) {
  resources(
    first: $first
    after: $after
    clientID: $clientID
    searchKeyword: $searchKeyword
  ) {
    edges {
      node {
        id
        name
        clientIDs
        resourceURI
        createdAt
        updatedAt
      }
      cursor
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    totalCount
  }
}
    `;

/**
 * __useResourcesQueryQuery__
 *
 * To run a query within a React component, call `useResourcesQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useResourcesQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourcesQueryQuery({
 *   variables: {
 *      first: // value for 'first'
 *      after: // value for 'after'
 *      clientID: // value for 'clientID'
 *      searchKeyword: // value for 'searchKeyword'
 *   },
 * });
 */
export function useResourcesQueryQuery(baseOptions?: Apollo.QueryHookOptions<ResourcesQueryQuery, ResourcesQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ResourcesQueryQuery, ResourcesQueryQueryVariables>(ResourcesQueryDocument, options);
      }
export function useResourcesQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ResourcesQueryQuery, ResourcesQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ResourcesQueryQuery, ResourcesQueryQueryVariables>(ResourcesQueryDocument, options);
        }
// @ts-ignore
export function useResourcesQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ResourcesQueryQuery, ResourcesQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ResourcesQueryQuery, ResourcesQueryQueryVariables>;
export function useResourcesQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ResourcesQueryQuery, ResourcesQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ResourcesQueryQuery | undefined, ResourcesQueryQueryVariables>;
export function useResourcesQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ResourcesQueryQuery, ResourcesQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ResourcesQueryQuery, ResourcesQueryQueryVariables>(ResourcesQueryDocument, options);
        }
export type ResourcesQueryQueryHookResult = ReturnType<typeof useResourcesQueryQuery>;
export type ResourcesQueryLazyQueryHookResult = ReturnType<typeof useResourcesQueryLazyQuery>;
export type ResourcesQuerySuspenseQueryHookResult = ReturnType<typeof useResourcesQuerySuspenseQuery>;
export type ResourcesQueryQueryResult = Apollo.QueryResult<ResourcesQueryQuery, ResourcesQueryQueryVariables>;