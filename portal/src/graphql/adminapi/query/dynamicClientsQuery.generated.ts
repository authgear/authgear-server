import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DynamicClientsQueryQueryVariables = Types.Exact<{
  first?: Types.InputMaybe<Types.Scalars['Int']['input']>;
  after?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type DynamicClientsQueryQuery = { __typename?: 'Query', dynamicClients?: { __typename?: 'OAuthClientConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'OAuthClientEdge', node?: { __typename?: 'OAuthClient', id: string, clientID: string, clientName?: string | null, name: string, kind: Types.OAuthClientKind, source: Types.OAuthClientSource, registeredAt?: any | null, applicationType?: string | null, redirectURIs: Array<string>, grantTypes: Array<string>, responseTypes: Array<string>, logoURI?: string | null, clientURI?: string | null, tosURI?: string | null, policyURI?: string | null } | null } | null> | null, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean } } | null };


export const DynamicClientsQueryDocument = gql`
    query dynamicClientsQuery($first: Int, $after: String) {
  dynamicClients(first: $first, after: $after) {
    edges {
      node {
        id
        clientID
        clientName
        name
        kind
        source
        registeredAt
        applicationType
        redirectURIs
        grantTypes
        responseTypes
        logoURI
        clientURI
        tosURI
        policyURI
      }
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
    `;

/**
 * __useDynamicClientsQueryQuery__
 *
 * To run a query within a React component, call `useDynamicClientsQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useDynamicClientsQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDynamicClientsQueryQuery({
 *   variables: {
 *      first: // value for 'first'
 *      after: // value for 'after'
 *   },
 * });
 */
export function useDynamicClientsQueryQuery(baseOptions?: Apollo.QueryHookOptions<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>(DynamicClientsQueryDocument, options);
      }
export function useDynamicClientsQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>(DynamicClientsQueryDocument, options);
        }
// @ts-ignore
export function useDynamicClientsQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>): Apollo.UseSuspenseQueryResult<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>;
export function useDynamicClientsQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>): Apollo.UseSuspenseQueryResult<DynamicClientsQueryQuery | undefined, DynamicClientsQueryQueryVariables>;
export function useDynamicClientsQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>(DynamicClientsQueryDocument, options);
        }
export type DynamicClientsQueryQueryHookResult = ReturnType<typeof useDynamicClientsQueryQuery>;
export type DynamicClientsQueryLazyQueryHookResult = ReturnType<typeof useDynamicClientsQueryLazyQuery>;
export type DynamicClientsQuerySuspenseQueryHookResult = ReturnType<typeof useDynamicClientsQuerySuspenseQuery>;
export type DynamicClientsQueryQueryResult = Apollo.QueryResult<DynamicClientsQueryQuery, DynamicClientsQueryQueryVariables>;