import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetMfaGracePeriodMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  endAt: Types.Scalars['DateTime']['input'];
}>;


export type SetMfaGracePeriodMutationMutation = { __typename?: 'Mutation', setMFAGracePeriod: { __typename?: 'SetMFAGracePeriodPayload', user: { __typename?: 'User', id: string, mfaGracePeriodEndAt?: any | null } } };


export const SetMfaGracePeriodMutationDocument = gql`
    mutation setMFAGracePeriodMutation($userID: ID!, $endAt: DateTime!) {
  setMFAGracePeriod(input: {userID: $userID, endAt: $endAt}) {
    user {
      id
      mfaGracePeriodEndAt
    }
  }
}
    `;
export type SetMfaGracePeriodMutationMutationFn = Apollo.MutationFunction<SetMfaGracePeriodMutationMutation, SetMfaGracePeriodMutationMutationVariables>;

/**
 * __useSetMfaGracePeriodMutationMutation__
 *
 * To run a mutation, you first call `useSetMfaGracePeriodMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetMfaGracePeriodMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setMfaGracePeriodMutationMutation, { data, loading, error }] = useSetMfaGracePeriodMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      endAt: // value for 'endAt'
 *   },
 * });
 */
export function useSetMfaGracePeriodMutationMutation(baseOptions?: Apollo.MutationHookOptions<SetMfaGracePeriodMutationMutation, SetMfaGracePeriodMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetMfaGracePeriodMutationMutation, SetMfaGracePeriodMutationMutationVariables>(SetMfaGracePeriodMutationDocument, options);
      }
export type SetMfaGracePeriodMutationMutationHookResult = ReturnType<typeof useSetMfaGracePeriodMutationMutation>;
export type SetMfaGracePeriodMutationMutationResult = Apollo.MutationResult<SetMfaGracePeriodMutationMutation>;
export type SetMfaGracePeriodMutationMutationOptions = Apollo.BaseMutationOptions<SetMfaGracePeriodMutationMutation, SetMfaGracePeriodMutationMutationVariables>;