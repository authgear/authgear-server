import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RolesListFragment = { __typename?: 'RoleConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'RoleEdge', cursor: string, node?: { __typename?: 'Role', id: string, createdAt: any, key: string, name?: string | null, description?: string | null } | null } | null> | null };

export type RolesListQueryQueryVariables = Types.Exact<{
  searchKeyword: Types.Scalars['String']['input'];
  excludedIDs?: Types.InputMaybe<Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input']>;
  pageSize: Types.Scalars['Int']['input'];
  cursor?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type RolesListQueryQuery = { __typename?: 'Query', roles?: { __typename?: 'RoleConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'RoleEdge', cursor: string, node?: { __typename?: 'Role', id: string, createdAt: any, key: string, name?: string | null, description?: string | null } | null } | null> | null } | null };

export const RolesListFragmentDoc = gql`
    fragment RolesList on RoleConnection {
  edges {
    node {
      id
      createdAt
      key
      name
      description
    }
    cursor
  }
  totalCount
}
    `;
export const RolesListQueryDocument = gql`
    query RolesListQuery($searchKeyword: String!, $excludedIDs: [ID!], $pageSize: Int!, $cursor: String) {
  roles(
    first: $pageSize
    after: $cursor
    searchKeyword: $searchKeyword
    excludedIDs: $excludedIDs
  ) {
    ...RolesList
  }
}
    ${RolesListFragmentDoc}`;

/**
 * __useRolesListQueryQuery__
 *
 * To run a query within a React component, call `useRolesListQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useRolesListQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useRolesListQueryQuery({
 *   variables: {
 *      searchKeyword: // value for 'searchKeyword'
 *      excludedIDs: // value for 'excludedIDs'
 *      pageSize: // value for 'pageSize'
 *      cursor: // value for 'cursor'
 *   },
 * });
 */
export function useRolesListQueryQuery(baseOptions: Apollo.QueryHookOptions<RolesListQueryQuery, RolesListQueryQueryVariables> & ({ variables: RolesListQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<RolesListQueryQuery, RolesListQueryQueryVariables>(RolesListQueryDocument, options);
      }
export function useRolesListQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<RolesListQueryQuery, RolesListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<RolesListQueryQuery, RolesListQueryQueryVariables>(RolesListQueryDocument, options);
        }
// @ts-ignore
export function useRolesListQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<RolesListQueryQuery, RolesListQueryQueryVariables>): Apollo.UseSuspenseQueryResult<RolesListQueryQuery, RolesListQueryQueryVariables>;
export function useRolesListQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<RolesListQueryQuery, RolesListQueryQueryVariables>): Apollo.UseSuspenseQueryResult<RolesListQueryQuery | undefined, RolesListQueryQueryVariables>;
export function useRolesListQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<RolesListQueryQuery, RolesListQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<RolesListQueryQuery, RolesListQueryQueryVariables>(RolesListQueryDocument, options);
        }
export type RolesListQueryQueryHookResult = ReturnType<typeof useRolesListQueryQuery>;
export type RolesListQueryLazyQueryHookResult = ReturnType<typeof useRolesListQueryLazyQuery>;
export type RolesListQuerySuspenseQueryHookResult = ReturnType<typeof useRolesListQuerySuspenseQuery>;
export type RolesListQueryQueryResult = Apollo.QueryResult<RolesListQueryQuery, RolesListQueryQueryVariables>;