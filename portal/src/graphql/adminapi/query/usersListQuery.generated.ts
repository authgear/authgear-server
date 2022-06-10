import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UsersListFragment = { __typename?: 'UserConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, createdAt: any, lastLoginAt?: any | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, standardAttributes: any, formattedName?: string | null, endUserAccountID?: string | null } | null } | null> | null };

export type UsersListQueryQueryVariables = Types.Exact<{
  searchKeyword: Types.Scalars['String'];
  pageSize: Types.Scalars['Int'];
  cursor?: Types.InputMaybe<Types.Scalars['String']>;
  sortBy?: Types.InputMaybe<Types.UserSortBy>;
  sortDirection?: Types.InputMaybe<Types.SortDirection>;
}>;


export type UsersListQueryQuery = { __typename?: 'Query', users?: { __typename?: 'UserConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, createdAt: any, lastLoginAt?: any | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, standardAttributes: any, formattedName?: string | null, endUserAccountID?: string | null } | null } | null> | null } | null };

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
      standardAttributes
      formattedName
      endUserAccountID
    }
  }
  totalCount
}
    `;
export const UsersListQueryDocument = gql`
    query UsersListQuery($searchKeyword: String!, $pageSize: Int!, $cursor: String, $sortBy: UserSortBy, $sortDirection: SortDirection) {
  users(
    first: $pageSize
    after: $cursor
    searchKeyword: $searchKeyword
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
export type UsersListQueryQueryHookResult = ReturnType<typeof useUsersListQueryQuery>;
export type UsersListQueryLazyQueryHookResult = ReturnType<typeof useUsersListQueryLazyQuery>;
export type UsersListQueryQueryResult = Apollo.QueryResult<UsersListQueryQuery, UsersListQueryQueryVariables>;