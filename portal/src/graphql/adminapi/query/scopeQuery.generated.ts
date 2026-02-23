import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ScopeQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type ScopeQueryQuery = { __typename?: 'Query', node?: { __typename: 'AuditLog' } | { __typename: 'Authenticator' } | { __typename: 'Authorization' } | { __typename: 'Group' } | { __typename: 'Identity' } | { __typename: 'Resource' } | { __typename: 'Role' } | { __typename: 'Scope', id: string, scope: string, description?: string | null, resourceID: string, createdAt: any, updatedAt: any } | { __typename: 'Session' } | { __typename: 'User' } | null };


export const ScopeQueryDocument = gql`
    query ScopeQuery($id: ID!) {
  node(id: $id) {
    __typename
    ... on Scope {
      id
      scope
      description
      resourceID
      createdAt
      updatedAt
    }
  }
}
    `;

/**
 * __useScopeQueryQuery__
 *
 * To run a query within a React component, call `useScopeQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useScopeQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useScopeQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useScopeQueryQuery(baseOptions: Apollo.QueryHookOptions<ScopeQueryQuery, ScopeQueryQueryVariables> & ({ variables: ScopeQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ScopeQueryQuery, ScopeQueryQueryVariables>(ScopeQueryDocument, options);
      }
export function useScopeQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ScopeQueryQuery, ScopeQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ScopeQueryQuery, ScopeQueryQueryVariables>(ScopeQueryDocument, options);
        }
// @ts-ignore
export function useScopeQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ScopeQueryQuery, ScopeQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ScopeQueryQuery, ScopeQueryQueryVariables>;
export function useScopeQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ScopeQueryQuery, ScopeQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ScopeQueryQuery | undefined, ScopeQueryQueryVariables>;
export function useScopeQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ScopeQueryQuery, ScopeQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ScopeQueryQuery, ScopeQueryQueryVariables>(ScopeQueryDocument, options);
        }
export type ScopeQueryQueryHookResult = ReturnType<typeof useScopeQueryQuery>;
export type ScopeQueryLazyQueryHookResult = ReturnType<typeof useScopeQueryLazyQuery>;
export type ScopeQuerySuspenseQueryHookResult = ReturnType<typeof useScopeQuerySuspenseQuery>;
export type ScopeQueryQueryResult = Apollo.QueryResult<ScopeQueryQuery, ScopeQueryQueryVariables>;