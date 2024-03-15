import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GroupQueryNodeFragment = { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null, createdAt: any, updatedAt: any, roles?: { __typename?: 'RoleConnection', edges?: Array<{ __typename?: 'RoleEdge', node?: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } | null } | null> | null } | null, users?: { __typename?: 'UserConnection', edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, formattedName?: string | null } | null } | null> | null } | null };

export type GroupQueryQueryVariables = Types.Exact<{
  groupID: Types.Scalars['ID']['input'];
}>;


export type GroupQueryQuery = { __typename?: 'Query', node?: { __typename: 'AuditLog' } | { __typename: 'Authenticator' } | { __typename: 'Authorization' } | { __typename: 'Group', id: string, key: string, name?: string | null, description?: string | null, createdAt: any, updatedAt: any, roles?: { __typename?: 'RoleConnection', edges?: Array<{ __typename?: 'RoleEdge', node?: { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null } | null } | null> | null } | null, users?: { __typename?: 'UserConnection', edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, formattedName?: string | null } | null } | null> | null } | null } | { __typename: 'Identity' } | { __typename: 'Role' } | { __typename: 'Session' } | { __typename: 'User' } | null };

export const GroupQueryNodeFragmentDoc = gql`
    fragment GroupQueryNode on Group {
  id
  key
  name
  description
  roles {
    edges {
      node {
        id
        key
        name
        description
      }
    }
  }
  users {
    edges {
      node {
        id
        formattedName
      }
    }
  }
  createdAt
  updatedAt
}
    `;
export const GroupQueryDocument = gql`
    query groupQuery($groupID: ID!) {
  node(id: $groupID) {
    __typename
    ...GroupQueryNode
  }
}
    ${GroupQueryNodeFragmentDoc}`;

/**
 * __useGroupQueryQuery__
 *
 * To run a query within a React component, call `useGroupQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useGroupQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGroupQueryQuery({
 *   variables: {
 *      groupID: // value for 'groupID'
 *   },
 * });
 */
export function useGroupQueryQuery(baseOptions: Apollo.QueryHookOptions<GroupQueryQuery, GroupQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GroupQueryQuery, GroupQueryQueryVariables>(GroupQueryDocument, options);
      }
export function useGroupQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GroupQueryQuery, GroupQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GroupQueryQuery, GroupQueryQueryVariables>(GroupQueryDocument, options);
        }
export function useGroupQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<GroupQueryQuery, GroupQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GroupQueryQuery, GroupQueryQueryVariables>(GroupQueryDocument, options);
        }
export type GroupQueryQueryHookResult = ReturnType<typeof useGroupQueryQuery>;
export type GroupQueryLazyQueryHookResult = ReturnType<typeof useGroupQueryLazyQuery>;
export type GroupQuerySuspenseQueryHookResult = ReturnType<typeof useGroupQuerySuspenseQuery>;
export type GroupQueryQueryResult = Apollo.QueryResult<GroupQueryQuery, GroupQueryQueryVariables>;