import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GetResourcesByClientIdQueryVariables = Types.Exact<{
  clientID: Types.Scalars['String']['input'];
  first?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  after?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type GetResourcesByClientIdQuery = { __typename?: 'Query', resources?: { __typename?: 'ResourceConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'ResourceEdge', cursor: string, node?: { __typename?: 'Resource', id: string, name?: string | null, resourceURI: string } | null } | null> | null, pageInfo: { __typename?: 'PageInfo', endCursor?: string | null, hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null } } | null };


export const GetResourcesByClientIdDocument = gql`
    query GetResourcesByClientID($clientID: String!, $first: Int, $after: String) {
  resources(clientID: $clientID, first: $first, after: $after) {
    edges {
      cursor
      node {
        id
        name
        resourceURI
      }
    }
    pageInfo {
      endCursor
      hasNextPage
      hasPreviousPage
      startCursor
    }
    totalCount
  }
}
    `;

/**
 * __useGetResourcesByClientIdQuery__
 *
 * To run a query within a React component, call `useGetResourcesByClientIdQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetResourcesByClientIdQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetResourcesByClientIdQuery({
 *   variables: {
 *      clientID: // value for 'clientID'
 *      first: // value for 'first'
 *      after: // value for 'after'
 *   },
 * });
 */
export function useGetResourcesByClientIdQuery(baseOptions: Apollo.QueryHookOptions<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables> & ({ variables: GetResourcesByClientIdQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>(GetResourcesByClientIdDocument, options);
      }
export function useGetResourcesByClientIdLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>(GetResourcesByClientIdDocument, options);
        }
// @ts-ignore
export function useGetResourcesByClientIdSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>): Apollo.UseSuspenseQueryResult<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>;
export function useGetResourcesByClientIdSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>): Apollo.UseSuspenseQueryResult<GetResourcesByClientIdQuery | undefined, GetResourcesByClientIdQueryVariables>;
export function useGetResourcesByClientIdSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>(GetResourcesByClientIdDocument, options);
        }
export type GetResourcesByClientIdQueryHookResult = ReturnType<typeof useGetResourcesByClientIdQuery>;
export type GetResourcesByClientIdLazyQueryHookResult = ReturnType<typeof useGetResourcesByClientIdLazyQuery>;
export type GetResourcesByClientIdSuspenseQueryHookResult = ReturnType<typeof useGetResourcesByClientIdSuspenseQuery>;
export type GetResourcesByClientIdQueryResult = Apollo.QueryResult<GetResourcesByClientIdQuery, GetResourcesByClientIdQueryVariables>;