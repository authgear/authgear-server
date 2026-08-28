import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RevokeInitialAccessTokenMutationMutationVariables = Types.Exact<{
  input: Types.RevokeInitialAccessTokenInput;
}>;


export type RevokeInitialAccessTokenMutationMutation = { __typename?: 'Mutation', revokeInitialAccessToken: { __typename?: 'RevokeInitialAccessTokenPayload', ok?: boolean | null } };


export const RevokeInitialAccessTokenMutationDocument = gql`
    mutation RevokeInitialAccessTokenMutation($input: RevokeInitialAccessTokenInput!) {
  revokeInitialAccessToken(input: $input) {
    ok
  }
}
    `;
export type RevokeInitialAccessTokenMutationMutationFn = Apollo.MutationFunction<RevokeInitialAccessTokenMutationMutation, RevokeInitialAccessTokenMutationMutationVariables>;

/**
 * __useRevokeInitialAccessTokenMutationMutation__
 *
 * To run a mutation, you first call `useRevokeInitialAccessTokenMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRevokeInitialAccessTokenMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [revokeInitialAccessTokenMutationMutation, { data, loading, error }] = useRevokeInitialAccessTokenMutationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useRevokeInitialAccessTokenMutationMutation(baseOptions?: Apollo.MutationHookOptions<RevokeInitialAccessTokenMutationMutation, RevokeInitialAccessTokenMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RevokeInitialAccessTokenMutationMutation, RevokeInitialAccessTokenMutationMutationVariables>(RevokeInitialAccessTokenMutationDocument, options);
      }
export type RevokeInitialAccessTokenMutationMutationHookResult = ReturnType<typeof useRevokeInitialAccessTokenMutationMutation>;
export type RevokeInitialAccessTokenMutationMutationResult = Apollo.MutationResult<RevokeInitialAccessTokenMutationMutation>;
export type RevokeInitialAccessTokenMutationMutationOptions = Apollo.BaseMutationOptions<RevokeInitialAccessTokenMutationMutation, RevokeInitialAccessTokenMutationMutationVariables>;