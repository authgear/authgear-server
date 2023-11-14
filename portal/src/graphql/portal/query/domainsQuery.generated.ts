import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DomainsQueryQueryVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
}>;


export type DomainsQueryQuery = { __typename?: 'Query', node?: { __typename?: 'App', id: string, domains: Array<{ __typename?: 'Domain', id: string, createdAt: any, apexDomain: string, domain: string, cookieDomain: string, isCustom: boolean, isVerified: boolean, verificationDNSRecord: string }> } | { __typename?: 'User' } | { __typename?: 'Viewer' } | null };


export const DomainsQueryDocument = gql`
    query domainsQuery($appID: ID!) {
  node(id: $appID) {
    ... on App {
      id
      domains {
        id
        createdAt
        apexDomain
        domain
        cookieDomain
        isCustom
        isVerified
        verificationDNSRecord
      }
    }
  }
}
    `;

/**
 * __useDomainsQueryQuery__
 *
 * To run a query within a React component, call `useDomainsQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useDomainsQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDomainsQueryQuery({
 *   variables: {
 *      appID: // value for 'appID'
 *   },
 * });
 */
export function useDomainsQueryQuery(baseOptions: Apollo.QueryHookOptions<DomainsQueryQuery, DomainsQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DomainsQueryQuery, DomainsQueryQueryVariables>(DomainsQueryDocument, options);
      }
export function useDomainsQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DomainsQueryQuery, DomainsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DomainsQueryQuery, DomainsQueryQueryVariables>(DomainsQueryDocument, options);
        }
export function useDomainsQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<DomainsQueryQuery, DomainsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<DomainsQueryQuery, DomainsQueryQueryVariables>(DomainsQueryDocument, options);
        }
export type DomainsQueryQueryHookResult = ReturnType<typeof useDomainsQueryQuery>;
export type DomainsQueryLazyQueryHookResult = ReturnType<typeof useDomainsQueryLazyQuery>;
export type DomainsQuerySuspenseQueryHookResult = ReturnType<typeof useDomainsQuerySuspenseQuery>;
export type DomainsQueryQueryResult = Apollo.QueryResult<DomainsQueryQuery, DomainsQueryQueryVariables>;