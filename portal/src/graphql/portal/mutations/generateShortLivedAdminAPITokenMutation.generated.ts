import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GenerateShortLivedAdminApiTokenMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  appSecretVisitToken: Types.Scalars['String']['input'];
}>;


export type GenerateShortLivedAdminApiTokenMutation = { __typename?: 'Mutation', generateShortLivedAdminAPIToken?: { __typename?: 'GenerateShortLivedAdminAPITokenPayload', token: string } | null };


export const GenerateShortLivedAdminApiTokenDocument = gql`
    mutation generateShortLivedAdminAPIToken($appID: ID!, $appSecretVisitToken: String!) {
  generateShortLivedAdminAPIToken(
    input: {appID: $appID, appSecretVisitToken: $appSecretVisitToken}
  ) {
    token
  }
}
    `;
export type GenerateShortLivedAdminApiTokenMutationFn = Apollo.MutationFunction<GenerateShortLivedAdminApiTokenMutation, GenerateShortLivedAdminApiTokenMutationVariables>;

/**
 * __useGenerateShortLivedAdminApiTokenMutation__
 *
 * To run a mutation, you first call `useGenerateShortLivedAdminApiTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useGenerateShortLivedAdminApiTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [generateShortLivedAdminApiTokenMutation, { data, loading, error }] = useGenerateShortLivedAdminApiTokenMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      appSecretVisitToken: // value for 'appSecretVisitToken'
 *   },
 * });
 */
export function useGenerateShortLivedAdminApiTokenMutation(baseOptions?: Apollo.MutationHookOptions<GenerateShortLivedAdminApiTokenMutation, GenerateShortLivedAdminApiTokenMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<GenerateShortLivedAdminApiTokenMutation, GenerateShortLivedAdminApiTokenMutationVariables>(GenerateShortLivedAdminApiTokenDocument, options);
      }
export type GenerateShortLivedAdminApiTokenMutationHookResult = ReturnType<typeof useGenerateShortLivedAdminApiTokenMutation>;
export type GenerateShortLivedAdminApiTokenMutationResult = Apollo.MutationResult<GenerateShortLivedAdminApiTokenMutation>;
export type GenerateShortLivedAdminApiTokenMutationOptions = Apollo.BaseMutationOptions<GenerateShortLivedAdminApiTokenMutation, GenerateShortLivedAdminApiTokenMutationVariables>;