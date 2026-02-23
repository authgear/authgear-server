import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ResourceScopesQueryQueryVariables = Types.Exact<{
  resourceID: Types.Scalars['ID']['input'];
  first?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  after?: Types.InputMaybe<Types.Scalars['String']['input']>;
  searchKeyword?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type ResourceScopesQueryQuery = { __typename?: 'Query', node?: { __typename?: 'AuditLog' } | { __typename?: 'Authenticator' } | { __typename?: 'Authorization' } | { __typename?: 'Group' } | { __typename?: 'Identity' } | { __typename?: 'Resource', id: string, resourceURI: string, name?: string | null, scopes?: { __typename?: 'ScopeConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'ScopeEdge', node?: { __typename?: 'Scope', id: string, scope: string, description?: string | null, createdAt: any, updatedAt: any } | null } | null> | null, pageInfo: { __typename?: 'PageInfo', endCursor?: string | null, hasNextPage: boolean } } | null } | { __typename?: 'Role' } | { __typename?: 'Scope' } | { __typename?: 'Session' } | { __typename?: 'User' } | null };


export const ResourceScopesQueryDocument = gql`
    query ResourceScopesQuery($resourceID: ID!, $first: Int, $after: String, $searchKeyword: String) {
  node(id: $resourceID) {
    ... on Resource {
      id
      resourceURI
      name
      scopes(first: $first, after: $after, searchKeyword: $searchKeyword) {
        edges {
          node {
            id
            scope
            description
            createdAt
            updatedAt
          }
        }
        pageInfo {
          endCursor
          hasNextPage
        }
        totalCount
      }
    }
  }
}
    `;

/**
 * __useResourceScopesQueryQuery__
 *
 * To run a query within a React component, call `useResourceScopesQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useResourceScopesQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourceScopesQueryQuery({
 *   variables: {
 *      resourceID: // value for 'resourceID'
 *      first: // value for 'first'
 *      after: // value for 'after'
 *      searchKeyword: // value for 'searchKeyword'
 *   },
 * });
 */
export function useResourceScopesQueryQuery(baseOptions: Apollo.QueryHookOptions<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables> & ({ variables: ResourceScopesQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>(ResourceScopesQueryDocument, options);
      }
export function useResourceScopesQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>(ResourceScopesQueryDocument, options);
        }
// @ts-ignore
export function useResourceScopesQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>;
export function useResourceScopesQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ResourceScopesQueryQuery | undefined, ResourceScopesQueryQueryVariables>;
export function useResourceScopesQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>(ResourceScopesQueryDocument, options);
        }
export type ResourceScopesQueryQueryHookResult = ReturnType<typeof useResourceScopesQueryQuery>;
export type ResourceScopesQueryLazyQueryHookResult = ReturnType<typeof useResourceScopesQueryLazyQuery>;
export type ResourceScopesQuerySuspenseQueryHookResult = ReturnType<typeof useResourceScopesQuerySuspenseQuery>;
export type ResourceScopesQueryQueryResult = Apollo.QueryResult<ResourceScopesQueryQuery, ResourceScopesQueryQueryVariables>;