import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UsersListFragment = { __typename?: 'UserConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'UserEdge', cursor: string, node?: { __typename?: 'User', id: string, createdAt: any, lastLoginAt?: any | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, standardAttributes: any, formattedName?: string | null, endUserAccountID?: string | null } | null } | null> | null };

export type UsersListQueryQueryVariables = Types.Exact<{
  searchKeyword: Types.Scalars['String']['input'];
  pageSize: Types.Scalars['Int']['input'];
  groupKeys?: Types.InputMaybe<Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input']>;
  roleKeys?: Types.InputMaybe<Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input']>;
  cursor?: Types.InputMaybe<Types.Scalars['String']['input']>;
  sortBy?: Types.InputMaybe<Types.UserSortBy>;
  sortDirection?: Types.InputMaybe<Types.SortDirection>;
}>;


export type UsersListQueryQuery = { __typename?: 'Query', users?: { __typename?: 'UserConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'UserEdge', cursor: string, node?: { __typename?: 'User', id: string, createdAt: any, lastLoginAt?: any | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, standardAttributes: any, formattedName?: string | null, endUserAccountID?: string | null } | null } | null> | null } | null };

export const UsersListFragmentDoc = gql`
    fragment UsersList on UserConnection {
  edges {
    node {
      id
      createdAt
      lastLoginAt
      isAnonymous
      isDisabled
      disableReason
      isDeactivated
      deleteAt
      isAnonymized
      anonymizeAt
      standardAttributes
      formattedName
      endUserAccountID
    }
    cursor
  }
  totalCount
}
    `;
export const UsersListQueryDocument = gql`
    query UsersListQuery($searchKeyword: String!, $pageSize: Int!, $groupKeys: [String!], $roleKeys: [String!], $cursor: String, $sortBy: UserSortBy, $sortDirection: SortDirection) {
  users(
    first: $pageSize
    after: $cursor
    searchKeyword: $searchKeyword
    groupKeys: $groupKeys
    roleKeys: $roleKeys
    sortBy: $sortBy
    sortDirection: $sortDirection
  ) {
    ...UsersList
  }
}
    ${UsersListFragmentDoc}`;

/**
 * __useUsersListQueryQuery__
 *
 * To run a query within a React component, call `useUsersListQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsersListQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsersListQueryQuery({
 *   variables: {
 *      searchKeyword: // value for 'searchKeyword'
 *      pageSize: // value for 'pageSize'
 *      groupKeys: // value for 'groupKeys'
 *      roleKeys: // value for 'roleKeys'
 *      cursor: // value for 'cursor'
 *      sortBy: // value for 'sortBy'
 *      sortDirection: // value for 'sortDirection'
 *   },
 * });
 */
export function useUsersListQueryQuery(baseOptions: Apollo.QueryHookOptions<UsersListQueryQuery, UsersListQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<UsersListQueryQuery, UsersListQueryQueryVariables>(UsersListQueryDocument, options);
      }
export function useUsersListQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<UsersListQueryQuery, UsersListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<UsersListQueryQuery, UsersListQueryQueryVariables>(UsersListQueryDocument, options);
        }
export function useUsersListQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<UsersListQueryQuery, UsersListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<UsersListQueryQuery, UsersListQueryQueryVariables>(UsersListQueryDocument, options);
        }
export type UsersListQueryQueryHookResult = ReturnType<typeof useUsersListQueryQuery>;
export type UsersListQueryLazyQueryHookResult = ReturnType<typeof useUsersListQueryLazyQuery>;
export type UsersListQuerySuspenseQueryHookResult = ReturnType<typeof useUsersListQuerySuspenseQuery>;
export type UsersListQueryQueryResult = Apollo.QueryResult<UsersListQueryQuery, UsersListQueryQueryVariables>;