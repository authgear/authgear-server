import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type NftCollectionsQueryQueryVariables = Types.Exact<{
  appID: Types.Scalars['ID'];
}>;


export type NftCollectionsQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, nftCollections: Array<{ __typename?: 'NFTCollection', name: string, blockchain: string, network: string, contractAddress: string, blockHeight: number, totalSupply: number, tokenType: string, createdAt: any }> } | { __typename: 'User' } | null };


export const NftCollectionsQueryDocument = gql`
    query nftCollectionsQuery($appID: ID!) {
  node(id: $appID) {
    __typename
    ... on App {
      id
      nftCollections {
        name
        blockchain
        network
        contractAddress
        blockHeight
        totalSupply
        tokenType
        createdAt
      }
    }
  }
}
    `;

/**
 * __useNftCollectionsQueryQuery__
 *
 * To run a query within a React component, call `useNftCollectionsQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useNftCollectionsQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useNftCollectionsQueryQuery({
 *   variables: {
 *      appID: // value for 'appID'
 *   },
 * });
 */
export function useNftCollectionsQueryQuery(baseOptions: Apollo.QueryHookOptions<NftCollectionsQueryQuery, NftCollectionsQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<NftCollectionsQueryQuery, NftCollectionsQueryQueryVariables>(NftCollectionsQueryDocument, options);
      }
export function useNftCollectionsQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<NftCollectionsQueryQuery, NftCollectionsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<NftCollectionsQueryQuery, NftCollectionsQueryQueryVariables>(NftCollectionsQueryDocument, options);
        }
export type NftCollectionsQueryQueryHookResult = ReturnType<typeof useNftCollectionsQueryQuery>;
export type NftCollectionsQueryLazyQueryHookResult = ReturnType<typeof useNftCollectionsQueryLazyQuery>;
export type NftCollectionsQueryQueryResult = Apollo.QueryResult<NftCollectionsQueryQuery, NftCollectionsQueryQueryVariables>;