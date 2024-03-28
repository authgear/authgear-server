import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GroupsListFragment = { __typename?: 'GroupConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'GroupEdge', cursor: string, node?: { __typename?: 'Group', id: string, createdAt: any, key: string, name?: string | null, description?: string | null } | null } | null> | null };

export type GroupsListQueryQueryVariables = Types.Exact<{
  searchKeyword: Types.Scalars['String']['input'];
  excludedIDs?: Types.InputMaybe<Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input']>;
  pageSize: Types.Scalars['Int']['input'];
  cursor?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type GroupsListQueryQuery = { __typename?: 'Query', groups?: { __typename?: 'GroupConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'GroupEdge', cursor: string, node?: { __typename?: 'Group', id: string, createdAt: any, key: string, name?: string | null, description?: string | null } | null } | null> | null } | null };

export const GroupsListFragmentDoc = gql`
    fragment GroupsList on GroupConnection {
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
export const GroupsListQueryDocument = gql`
    query GroupsListQuery($searchKeyword: String!, $excludedIDs: [ID!], $pageSize: Int!, $cursor: String) {
  groups(
    first: $pageSize
    after: $cursor
    searchKeyword: $searchKeyword
    excludedIDs: $excludedIDs
  ) {
    ...GroupsList
  }
}
    ${GroupsListFragmentDoc}`;

/**
 * __useGroupsListQueryQuery__
 *
 * To run a query within a React component, call `useGroupsListQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useGroupsListQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGroupsListQueryQuery({
 *   variables: {
 *      searchKeyword: // value for 'searchKeyword'
 *      excludedIDs: // value for 'excludedIDs'
 *      pageSize: // value for 'pageSize'
 *      cursor: // value for 'cursor'
 *   },
 * });
 */
export function useGroupsListQueryQuery(baseOptions: Apollo.QueryHookOptions<GroupsListQueryQuery, GroupsListQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GroupsListQueryQuery, GroupsListQueryQueryVariables>(GroupsListQueryDocument, options);
      }
export function useGroupsListQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GroupsListQueryQuery, GroupsListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GroupsListQueryQuery, GroupsListQueryQueryVariables>(GroupsListQueryDocument, options);
        }
export function useGroupsListQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<GroupsListQueryQuery, GroupsListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GroupsListQueryQuery, GroupsListQueryQueryVariables>(GroupsListQueryDocument, options);
        }
export type GroupsListQueryQueryHookResult = ReturnType<typeof useGroupsListQueryQuery>;
export type GroupsListQueryLazyQueryHookResult = ReturnType<typeof useGroupsListQueryLazyQuery>;
export type GroupsListQuerySuspenseQueryHookResult = ReturnType<typeof useGroupsListQuerySuspenseQuery>;
export type GroupsListQueryQueryResult = Apollo.QueryResult<GroupsListQueryQuery, GroupsListQueryQueryVariables>;