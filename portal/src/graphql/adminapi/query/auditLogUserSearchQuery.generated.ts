import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AuditLogUserSearchQueryQueryVariables = Types.Exact<{
  searchKeyword: Types.Scalars['String']['input'];
}>;


export type AuditLogUserSearchQueryQuery = { __typename?: 'Query', users?: { __typename?: 'UserConnection', edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string } | null } | null> | null } | null };


export const AuditLogUserSearchQueryDocument = gql`
    query AuditLogUserSearchQuery($searchKeyword: String!) {
  users(first: 50, searchKeyword: $searchKeyword) {
    edges {
      node {
        id
      }
    }
  }
}
    `;

/**
 * __useAuditLogUserSearchQueryQuery__
 *
 * To run a query within a React component, call `useAuditLogUserSearchQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAuditLogUserSearchQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAuditLogUserSearchQueryQuery({
 *   variables: {
 *      searchKeyword: // value for 'searchKeyword'
 *   },
 * });
 */
export function useAuditLogUserSearchQueryQuery(baseOptions: Apollo.QueryHookOptions<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables> & ({ variables: AuditLogUserSearchQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>(AuditLogUserSearchQueryDocument, options);
      }
export function useAuditLogUserSearchQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>(AuditLogUserSearchQueryDocument, options);
        }
// @ts-ignore
export function useAuditLogUserSearchQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>;
export function useAuditLogUserSearchQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AuditLogUserSearchQueryQuery | undefined, AuditLogUserSearchQueryQueryVariables>;
export function useAuditLogUserSearchQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>(AuditLogUserSearchQueryDocument, options);
        }
export type AuditLogUserSearchQueryQueryHookResult = ReturnType<typeof useAuditLogUserSearchQueryQuery>;
export type AuditLogUserSearchQueryLazyQueryHookResult = ReturnType<typeof useAuditLogUserSearchQueryLazyQuery>;
export type AuditLogUserSearchQuerySuspenseQueryHookResult = ReturnType<typeof useAuditLogUserSearchQuerySuspenseQuery>;
export type AuditLogUserSearchQueryQueryResult = Apollo.QueryResult<AuditLogUserSearchQueryQuery, AuditLogUserSearchQueryQueryVariables>;