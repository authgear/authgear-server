import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RoleQueryNodeFragment = { __typename?: 'Role', id: string, key: string, name?: string | null, description?: string | null, createdAt: any, updatedAt: any, groups?: { __typename?: 'GroupConnection', edges?: Array<{ __typename?: 'GroupEdge', node?: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } | null } | null> | null } | null, users?: { __typename?: 'UserConnection', edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, formattedName?: string | null } | null } | null> | null } | null };

export type RoleQueryQueryVariables = Types.Exact<{
  roleID: Types.Scalars['ID']['input'];
}>;


export type RoleQueryQuery = { __typename?: 'Query', node?: { __typename: 'AuditLog' } | { __typename: 'Authenticator' } | { __typename: 'Authorization' } | { __typename: 'Group' } | { __typename: 'Identity' } | { __typename: 'Resource' } | { __typename: 'Role', id: string, key: string, name?: string | null, description?: string | null, createdAt: any, updatedAt: any, groups?: { __typename?: 'GroupConnection', edges?: Array<{ __typename?: 'GroupEdge', node?: { __typename?: 'Group', id: string, key: string, name?: string | null, description?: string | null } | null } | null> | null } | null, users?: { __typename?: 'UserConnection', edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, formattedName?: string | null } | null } | null> | null } | null } | { __typename: 'Scope' } | { __typename: 'Session' } | { __typename: 'User' } | null };

export const RoleQueryNodeFragmentDoc = gql`
    fragment RoleQueryNode on Role {
  id
  key
  name
  description
  groups {
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
export const RoleQueryDocument = gql`
    query roleQuery($roleID: ID!) {
  node(id: $roleID) {
    __typename
    ...RoleQueryNode
  }
}
    ${RoleQueryNodeFragmentDoc}`;

/**
 * __useRoleQueryQuery__
 *
 * To run a query within a React component, call `useRoleQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useRoleQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useRoleQueryQuery({
 *   variables: {
 *      roleID: // value for 'roleID'
 *   },
 * });
 */
export function useRoleQueryQuery(baseOptions: Apollo.QueryHookOptions<RoleQueryQuery, RoleQueryQueryVariables> & ({ variables: RoleQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<RoleQueryQuery, RoleQueryQueryVariables>(RoleQueryDocument, options);
      }
export function useRoleQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<RoleQueryQuery, RoleQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<RoleQueryQuery, RoleQueryQueryVariables>(RoleQueryDocument, options);
        }
// @ts-ignore
export function useRoleQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<RoleQueryQuery, RoleQueryQueryVariables>): Apollo.UseSuspenseQueryResult<RoleQueryQuery, RoleQueryQueryVariables>;
export function useRoleQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<RoleQueryQuery, RoleQueryQueryVariables>): Apollo.UseSuspenseQueryResult<RoleQueryQuery | undefined, RoleQueryQueryVariables>;
export function useRoleQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<RoleQueryQuery, RoleQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<RoleQueryQuery, RoleQueryQueryVariables>(RoleQueryDocument, options);
        }
export type RoleQueryQueryHookResult = ReturnType<typeof useRoleQueryQuery>;
export type RoleQueryLazyQueryHookResult = ReturnType<typeof useRoleQueryLazyQuery>;
export type RoleQuerySuspenseQueryHookResult = ReturnType<typeof useRoleQuerySuspenseQuery>;
export type RoleQueryQueryResult = Apollo.QueryResult<RoleQueryQuery, RoleQueryQueryVariables>;