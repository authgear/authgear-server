import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ResourceQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type ResourceQueryQuery = { __typename?: 'Query', node?: { __typename?: 'AuditLog' } | { __typename?: 'Authenticator' } | { __typename?: 'Authorization' } | { __typename?: 'Group' } | { __typename?: 'Identity' } | { __typename?: 'Resource', id: string, name?: string | null, resourceURI: string, clientIDs: Array<string>, createdAt: any, updatedAt: any } | { __typename?: 'Role' } | { __typename?: 'Scope' } | { __typename?: 'Session' } | { __typename?: 'User' } | null };


export const ResourceQueryDocument = gql`
    query ResourceQuery($id: ID!) {
  node(id: $id) {
    ... on Resource {
      id
      name
      resourceURI
      clientIDs
      createdAt
      updatedAt
    }
  }
}
    `;

/**
 * __useResourceQueryQuery__
 *
 * To run a query within a React component, call `useResourceQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useResourceQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourceQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useResourceQueryQuery(baseOptions: Apollo.QueryHookOptions<ResourceQueryQuery, ResourceQueryQueryVariables> & ({ variables: ResourceQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ResourceQueryQuery, ResourceQueryQueryVariables>(ResourceQueryDocument, options);
      }
export function useResourceQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ResourceQueryQuery, ResourceQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ResourceQueryQuery, ResourceQueryQueryVariables>(ResourceQueryDocument, options);
        }
// @ts-ignore
export function useResourceQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ResourceQueryQuery, ResourceQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ResourceQueryQuery, ResourceQueryQueryVariables>;
export function useResourceQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ResourceQueryQuery, ResourceQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ResourceQueryQuery | undefined, ResourceQueryQueryVariables>;
export function useResourceQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ResourceQueryQuery, ResourceQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ResourceQueryQuery, ResourceQueryQueryVariables>(ResourceQueryDocument, options);
        }
export type ResourceQueryQueryHookResult = ReturnType<typeof useResourceQueryQuery>;
export type ResourceQueryLazyQueryHookResult = ReturnType<typeof useResourceQueryLazyQuery>;
export type ResourceQuerySuspenseQueryHookResult = ReturnType<typeof useResourceQuerySuspenseQuery>;
export type ResourceQueryQueryResult = Apollo.QueryResult<ResourceQueryQuery, ResourceQueryQueryVariables>;