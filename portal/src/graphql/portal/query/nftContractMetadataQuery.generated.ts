import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type NftContractMetadataQueryQueryVariables = Types.Exact<{
  contractID: Types.Scalars['String'];
}>;


export type NftContractMetadataQueryQuery = { __typename?: 'Query', nftContractMetadata?: { __typename?: 'NFTCollection', name: string, blockchain: string, network: string, contractAddress: string, totalSupply?: string | null, tokenType: string, createdAt: any } | null };


export const NftContractMetadataQueryDocument = gql`
    query nftContractMetadataQuery($contractID: String!) {
  nftContractMetadata(contractID: $contractID) {
    name
    blockchain
    network
    contractAddress
    totalSupply
    tokenType
    createdAt
  }
}
    `;

/**
 * __useNftContractMetadataQueryQuery__
 *
 * To run a query within a React component, call `useNftContractMetadataQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useNftContractMetadataQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useNftContractMetadataQueryQuery({
 *   variables: {
 *      contractID: // value for 'contractID'
 *   },
 * });
 */
export function useNftContractMetadataQueryQuery(baseOptions: Apollo.QueryHookOptions<NftContractMetadataQueryQuery, NftContractMetadataQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<NftContractMetadataQueryQuery, NftContractMetadataQueryQueryVariables>(NftContractMetadataQueryDocument, options);
      }
export function useNftContractMetadataQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<NftContractMetadataQueryQuery, NftContractMetadataQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<NftContractMetadataQueryQuery, NftContractMetadataQueryQueryVariables>(NftContractMetadataQueryDocument, options);
        }
export type NftContractMetadataQueryQueryHookResult = ReturnType<typeof useNftContractMetadataQueryQuery>;
export type NftContractMetadataQueryLazyQueryHookResult = ReturnType<typeof useNftContractMetadataQueryLazyQuery>;
export type NftContractMetadataQueryQueryResult = Apollo.QueryResult<NftContractMetadataQueryQuery, NftContractMetadataQueryQueryVariables>;