import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ProbeNftCollectionMutationMutationVariables = Types.Exact<{
  contractID: Types.Scalars['String'];
}>;


export type ProbeNftCollectionMutationMutation = { __typename?: 'Mutation', probeNFTCollection: { __typename?: 'ProbeNFTCollectionsPayload', isLargeCollection: boolean } };


export const ProbeNftCollectionMutationDocument = gql`
    mutation probeNFTCollectionMutation($contractID: String!) {
  probeNFTCollection(input: {contractID: $contractID}) {
    isLargeCollection
  }
}
    `;
export type ProbeNftCollectionMutationMutationFn = Apollo.MutationFunction<ProbeNftCollectionMutationMutation, ProbeNftCollectionMutationMutationVariables>;

/**
 * __useProbeNftCollectionMutationMutation__
 *
 * To run a mutation, you first call `useProbeNftCollectionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useProbeNftCollectionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [probeNftCollectionMutationMutation, { data, loading, error }] = useProbeNftCollectionMutationMutation({
 *   variables: {
 *      contractID: // value for 'contractID'
 *   },
 * });
 */
export function useProbeNftCollectionMutationMutation(baseOptions?: Apollo.MutationHookOptions<ProbeNftCollectionMutationMutation, ProbeNftCollectionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ProbeNftCollectionMutationMutation, ProbeNftCollectionMutationMutationVariables>(ProbeNftCollectionMutationDocument, options);
      }
export type ProbeNftCollectionMutationMutationHookResult = ReturnType<typeof useProbeNftCollectionMutationMutation>;
export type ProbeNftCollectionMutationMutationResult = Apollo.MutationResult<ProbeNftCollectionMutationMutation>;
export type ProbeNftCollectionMutationMutationOptions = Apollo.BaseMutationOptions<ProbeNftCollectionMutationMutation, ProbeNftCollectionMutationMutationVariables>;