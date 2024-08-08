import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RemoveMfaGracePeriodMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type RemoveMfaGracePeriodMutationMutation = { __typename?: 'Mutation', removeMFAGracePeriod: { __typename?: 'removeMFAGracePeriodPayload', user: { __typename?: 'User', id: string, mfaGracePeriodEndAt?: any | null } } };


export const RemoveMfaGracePeriodMutationDocument = gql`
    mutation removeMFAGracePeriodMutation($userID: ID!) {
  removeMFAGracePeriod(input: {userID: $userID}) {
    user {
      id
      mfaGracePeriodEndAt
    }
  }
}
    `;
export type RemoveMfaGracePeriodMutationMutationFn = Apollo.MutationFunction<RemoveMfaGracePeriodMutationMutation, RemoveMfaGracePeriodMutationMutationVariables>;

/**
 * __useRemoveMfaGracePeriodMutationMutation__
 *
 * To run a mutation, you first call `useRemoveMfaGracePeriodMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveMfaGracePeriodMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeMfaGracePeriodMutationMutation, { data, loading, error }] = useRemoveMfaGracePeriodMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useRemoveMfaGracePeriodMutationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveMfaGracePeriodMutationMutation, RemoveMfaGracePeriodMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveMfaGracePeriodMutationMutation, RemoveMfaGracePeriodMutationMutationVariables>(RemoveMfaGracePeriodMutationDocument, options);
      }
export type RemoveMfaGracePeriodMutationMutationHookResult = ReturnType<typeof useRemoveMfaGracePeriodMutationMutation>;
export type RemoveMfaGracePeriodMutationMutationResult = Apollo.MutationResult<RemoveMfaGracePeriodMutationMutation>;
export type RemoveMfaGracePeriodMutationMutationOptions = Apollo.BaseMutationOptions<RemoveMfaGracePeriodMutationMutation, RemoveMfaGracePeriodMutationMutationVariables>;