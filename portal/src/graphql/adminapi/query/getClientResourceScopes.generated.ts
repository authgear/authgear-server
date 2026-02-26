import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GetClientResourceScopesQueryVariables = Types.Exact<{
  resourceID: Types.Scalars['ID']['input'];
  clientID: Types.Scalars['String']['input'];
  first?: Types.InputMaybe<Types.Scalars['Int']['input']>;
}>;


export type GetClientResourceScopesQuery = { __typename?: 'Query', node?: { __typename?: 'AuditLog' } | { __typename?: 'Authenticator' } | { __typename?: 'Authorization' } | { __typename?: 'Group' } | { __typename?: 'Identity' } | { __typename?: 'Resource', id: string, name?: string | null, resourceURI: string, createdAt: any, updatedAt: any, scopes?: { __typename?: 'ScopeConnection', edges?: Array<{ __typename?: 'ScopeEdge', node?: { __typename?: 'Scope', id: string, scope: string, resourceID: string, description?: string | null, createdAt: any, updatedAt: any } | null } | null> | null } | null } | { __typename?: 'Role' } | { __typename?: 'Scope' } | { __typename?: 'Session' } | { __typename?: 'User' } | null };


export const GetClientResourceScopesDocument = gql`
    query GetClientResourceScopes($resourceID: ID!, $clientID: String!, $first: Int) {
  node(id: $resourceID) {
    ... on Resource {
      id
      name
      resourceURI
      scopes(clientID: $clientID, first: $first) {
        edges {
          node {
            id
            scope
            resourceID
            description
            createdAt
            updatedAt
          }
        }
      }
      createdAt
      updatedAt
    }
  }
}
    `;

/**
 * __useGetClientResourceScopesQuery__
 *
 * To run a query within a React component, call `useGetClientResourceScopesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetClientResourceScopesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetClientResourceScopesQuery({
 *   variables: {
 *      resourceID: // value for 'resourceID'
 *      clientID: // value for 'clientID'
 *      first: // value for 'first'
 *   },
 * });
 */
export function useGetClientResourceScopesQuery(baseOptions: Apollo.QueryHookOptions<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables> & ({ variables: GetClientResourceScopesQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>(GetClientResourceScopesDocument, options);
      }
export function useGetClientResourceScopesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>(GetClientResourceScopesDocument, options);
        }
// @ts-ignore
export function useGetClientResourceScopesSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>): Apollo.UseSuspenseQueryResult<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>;
export function useGetClientResourceScopesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>): Apollo.UseSuspenseQueryResult<GetClientResourceScopesQuery | undefined, GetClientResourceScopesQueryVariables>;
export function useGetClientResourceScopesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>(GetClientResourceScopesDocument, options);
        }
export type GetClientResourceScopesQueryHookResult = ReturnType<typeof useGetClientResourceScopesQuery>;
export type GetClientResourceScopesLazyQueryHookResult = ReturnType<typeof useGetClientResourceScopesLazyQuery>;
export type GetClientResourceScopesSuspenseQueryHookResult = ReturnType<typeof useGetClientResourceScopesSuspenseQuery>;
export type GetClientResourceScopesQueryResult = Apollo.QueryResult<GetClientResourceScopesQuery, GetClientResourceScopesQueryVariables>;