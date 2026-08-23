import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateInitialAccessTokenMutationMutationVariables = Types.Exact<{
  input: Types.CreateInitialAccessTokenInput;
}>;


export type CreateInitialAccessTokenMutationMutation = { __typename?: 'Mutation', createInitialAccessToken: { __typename?: 'CreateInitialAccessTokenPayload', token: string, initialAccessToken: { __typename?: 'InitialAccessToken', id: string, createdAt: any, expiresAt: any, type: Types.InitialAccessTokenType } } };


export const CreateInitialAccessTokenMutationDocument = gql`
    mutation CreateInitialAccessTokenMutation($input: CreateInitialAccessTokenInput!) {
  createInitialAccessToken(input: $input) {
    token
    initialAccessToken {
      id
      createdAt
      expiresAt
      type
    }
  }
}
    `;
export type CreateInitialAccessTokenMutationMutationFn = Apollo.MutationFunction<CreateInitialAccessTokenMutationMutation, CreateInitialAccessTokenMutationMutationVariables>;

/**
 * __useCreateInitialAccessTokenMutationMutation__
 *
 * To run a mutation, you first call `useCreateInitialAccessTokenMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateInitialAccessTokenMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createInitialAccessTokenMutationMutation, { data, loading, error }] = useCreateInitialAccessTokenMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateInitialAccessTokenMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateInitialAccessTokenMutationMutation, CreateInitialAccessTokenMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateInitialAccessTokenMutationMutation, CreateInitialAccessTokenMutationMutationVariables>(CreateInitialAccessTokenMutationDocument, options);
      }
export type CreateInitialAccessTokenMutationMutationHookResult = ReturnType<typeof useCreateInitialAccessTokenMutationMutation>;
export type CreateInitialAccessTokenMutationMutationResult = Apollo.MutationResult<CreateInitialAccessTokenMutationMutation>;
export type CreateInitialAccessTokenMutationMutationOptions = Apollo.BaseMutationOptions<CreateInitialAccessTokenMutationMutation, CreateInitialAccessTokenMutationMutationVariables>;