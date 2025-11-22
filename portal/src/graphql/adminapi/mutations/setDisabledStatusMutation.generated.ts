import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetDisabledStatusMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  isDisabled: Types.Scalars['Boolean']['input'];
  reason?: Types.InputMaybe<Types.Scalars['String']['input']>;
  temporarilyDisabledFrom?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  temporarilyDisabledUntil?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
}>;


export type SetDisabledStatusMutationMutation = { __typename?: 'Mutation', setDisabledStatus: { __typename?: 'SetDisabledStatusPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, temporarilyDisabledFrom?: any | null, temporarilyDisabledUntil?: any | null, accountValidFrom?: any | null, accountValidUntil?: any | null } } };


export const SetDisabledStatusMutationDocument = gql`
    mutation setDisabledStatusMutation($userID: ID!, $isDisabled: Boolean!, $reason: String, $temporarilyDisabledFrom: DateTime, $temporarilyDisabledUntil: DateTime) {
  setDisabledStatus(
    input: {userID: $userID, reason: $reason, isDisabled: $isDisabled, temporarilyDisabledFrom: $temporarilyDisabledFrom, temporarilyDisabledUntil: $temporarilyDisabledUntil}
  ) {
    user {
      id
      isDisabled
      disableReason
      isDeactivated
      deleteAt
      isAnonymized
      anonymizeAt
      temporarilyDisabledFrom
      temporarilyDisabledUntil
      accountValidFrom
      accountValidUntil
    }
  }
}
    `;
export type SetDisabledStatusMutationMutationFn = Apollo.MutationFunction<SetDisabledStatusMutationMutation, SetDisabledStatusMutationMutationVariables>;

/**
 * __useSetDisabledStatusMutationMutation__
 *
 * To run a mutation, you first call `useSetDisabledStatusMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetDisabledStatusMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setDisabledStatusMutationMutation, { data, loading, error }] = useSetDisabledStatusMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      isDisabled: // value for 'isDisabled'
 *      reason: // value for 'reason'
 *      temporarilyDisabledFrom: // value for 'temporarilyDisabledFrom'
 *      temporarilyDisabledUntil: // value for 'temporarilyDisabledUntil'
 *   },
 * });
 */
export function useSetDisabledStatusMutationMutation(baseOptions?: Apollo.MutationHookOptions<SetDisabledStatusMutationMutation, SetDisabledStatusMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetDisabledStatusMutationMutation, SetDisabledStatusMutationMutationVariables>(SetDisabledStatusMutationDocument, options);
      }
export type SetDisabledStatusMutationMutationHookResult = ReturnType<typeof useSetDisabledStatusMutationMutation>;
export type SetDisabledStatusMutationMutationResult = Apollo.MutationResult<SetDisabledStatusMutationMutation>;
export type SetDisabledStatusMutationMutationOptions = Apollo.BaseMutationOptions<SetDisabledStatusMutationMutation, SetDisabledStatusMutationMutationVariables>;