import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type GenerateAppSecretVisitTokenMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  secrets: Array<Types.AppSecretKey> | Types.AppSecretKey;
}>;


export type GenerateAppSecretVisitTokenMutationMutation = { __typename?: 'Mutation', generateAppSecretVisitToken: { __typename?: 'GenerateAppSecretVisitTokenPayloadPayload', token: string } };


export const GenerateAppSecretVisitTokenMutationDocument = gql`
    mutation generateAppSecretVisitTokenMutation($appID: ID!, $secrets: [AppSecretKey!]!) {
  generateAppSecretVisitToken(input: {id: $appID, secrets: $secrets}) {
    token
  }
}
    `;
export type GenerateAppSecretVisitTokenMutationMutationFn = Apollo.MutationFunction<GenerateAppSecretVisitTokenMutationMutation, GenerateAppSecretVisitTokenMutationMutationVariables>;

/**
 * __useGenerateAppSecretVisitTokenMutationMutation__
 *
 * To run a mutation, you first call `useGenerateAppSecretVisitTokenMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useGenerateAppSecretVisitTokenMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [generateAppSecretVisitTokenMutationMutation, { data, loading, error }] = useGenerateAppSecretVisitTokenMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      secrets: // value for 'secrets'
 *   },
 * });
 */
export function useGenerateAppSecretVisitTokenMutationMutation(baseOptions?: Apollo.MutationHookOptions<GenerateAppSecretVisitTokenMutationMutation, GenerateAppSecretVisitTokenMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<GenerateAppSecretVisitTokenMutationMutation, GenerateAppSecretVisitTokenMutationMutationVariables>(GenerateAppSecretVisitTokenMutationDocument, options);
      }
export type GenerateAppSecretVisitTokenMutationMutationHookResult = ReturnType<typeof useGenerateAppSecretVisitTokenMutationMutation>;
export type GenerateAppSecretVisitTokenMutationMutationResult = Apollo.MutationResult<GenerateAppSecretVisitTokenMutationMutation>;
export type GenerateAppSecretVisitTokenMutationMutationOptions = Apollo.BaseMutationOptions<GenerateAppSecretVisitTokenMutationMutation, GenerateAppSecretVisitTokenMutationMutationVariables>;