import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type WatchNftCollectionsMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['String'];
  contractIDs: Array<Types.Scalars['String']> | Types.Scalars['String'];
}>;


export type WatchNftCollectionsMutationMutation = { __typename?: 'Mutation', watchNFTCollections: { __typename?: 'WatchNFTCollectionsPayload', app: { __typename?: 'App', id: string, nftCollections: Array<{ __typename?: 'NFTCollection', name: string, blockchain: string, network: string, contractAddress: string, totalSupply: number, tokenType: string, createdAt: any }> } } };


export const WatchNftCollectionsMutationDocument = gql`
    mutation watchNFTCollectionsMutation($appID: String!, $contractIDs: [String!]!) {
  watchNFTCollections(input: {id: $appID, contractIDs: $contractIDs}) {
    app {
      id
      nftCollections {
        name
        blockchain
        network
        contractAddress
        totalSupply
        tokenType
        createdAt
      }
    }
  }
}
    `;
export type WatchNftCollectionsMutationMutationFn = Apollo.MutationFunction<WatchNftCollectionsMutationMutation, WatchNftCollectionsMutationMutationVariables>;

/**
 * __useWatchNftCollectionsMutationMutation__
 *
 * To run a mutation, you first call `useWatchNftCollectionsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useWatchNftCollectionsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [watchNftCollectionsMutationMutation, { data, loading, error }] = useWatchNftCollectionsMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      contractIDs: // value for 'contractIDs'
 *   },
 * });
 */
export function useWatchNftCollectionsMutationMutation(baseOptions?: Apollo.MutationHookOptions<WatchNftCollectionsMutationMutation, WatchNftCollectionsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<WatchNftCollectionsMutationMutation, WatchNftCollectionsMutationMutationVariables>(WatchNftCollectionsMutationDocument, options);
      }
export type WatchNftCollectionsMutationMutationHookResult = ReturnType<typeof useWatchNftCollectionsMutationMutation>;
export type WatchNftCollectionsMutationMutationResult = Apollo.MutationResult<WatchNftCollectionsMutationMutation>;
export type WatchNftCollectionsMutationMutationOptions = Apollo.BaseMutationOptions<WatchNftCollectionsMutationMutation, WatchNftCollectionsMutationMutationVariables>;