import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DynamicAccessResourcesQueryQueryVariables = Types.Exact<{
  first?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  scopesFirst?: Types.InputMaybe<Types.Scalars['Int']['input']>;
}>;


export type DynamicAccessResourcesQueryQuery = { __typename?: 'Query', resources?: { __typename?: 'ResourceConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'ResourceEdge', node?: { __typename?: 'Resource', id: string, name?: string | null, resourceURI: string, accessPolicy: { __typename?: 'AccessPolicy', allowDynamicThirdPartyClientAccess: boolean }, scopes?: { __typename?: 'ScopeConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'ScopeEdge', node?: { __typename?: 'Scope', id: string, scope: string, accessPolicy: { __typename?: 'AccessPolicy', allowDynamicThirdPartyClientAccess: boolean } } | null } | null> | null } | null } | null } | null> | null } | null };


export const DynamicAccessResourcesQueryDocument = gql`
    query dynamicAccessResourcesQuery($first: Int, $scopesFirst: Int) {
  resources(first: $first) {
    totalCount
    edges {
      node {
        id
        name
        resourceURI
        accessPolicy {
          allowDynamicThirdPartyClientAccess
        }
        scopes(first: $scopesFirst) {
          totalCount
          edges {
            node {
              id
              scope
              accessPolicy {
                allowDynamicThirdPartyClientAccess
              }
            }
          }
        }
      }
    }
  }
}
    `;

/**
 * __useDynamicAccessResourcesQueryQuery__
 *
 * To run a query within a React component, call `useDynamicAccessResourcesQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useDynamicAccessResourcesQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDynamicAccessResourcesQueryQuery({
 *   variables: {
 *      first: // value for 'first'
 *      scopesFirst: // value for 'scopesFirst'
 *   },
 * });
 */
export function useDynamicAccessResourcesQueryQuery(baseOptions?: Apollo.QueryHookOptions<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>(DynamicAccessResourcesQueryDocument, options);
      }
export function useDynamicAccessResourcesQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>(DynamicAccessResourcesQueryDocument, options);
        }
// @ts-ignore
export function useDynamicAccessResourcesQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>): Apollo.UseSuspenseQueryResult<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>;
export function useDynamicAccessResourcesQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>): Apollo.UseSuspenseQueryResult<DynamicAccessResourcesQueryQuery | undefined, DynamicAccessResourcesQueryQueryVariables>;
export function useDynamicAccessResourcesQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>(DynamicAccessResourcesQueryDocument, options);
        }
export type DynamicAccessResourcesQueryQueryHookResult = ReturnType<typeof useDynamicAccessResourcesQueryQuery>;
export type DynamicAccessResourcesQueryLazyQueryHookResult = ReturnType<typeof useDynamicAccessResourcesQueryLazyQuery>;
export type DynamicAccessResourcesQuerySuspenseQueryHookResult = ReturnType<typeof useDynamicAccessResourcesQuerySuspenseQuery>;
export type DynamicAccessResourcesQueryQueryResult = Apollo.QueryResult<DynamicAccessResourcesQueryQuery, DynamicAccessResourcesQueryQueryVariables>;