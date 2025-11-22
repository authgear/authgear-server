import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetAccountValidPeriodMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  accountValidFrom?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  accountValidUntil?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
}>;


export type SetAccountValidPeriodMutationMutation = { __typename?: 'Mutation', setAccountValidPeriod: { __typename?: 'SetAccountValidPeriodPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null, temporarilyDisabledFrom?: any | null, temporarilyDisabledUntil?: any | null, accountValidFrom?: any | null, accountValidUntil?: any | null } } };


export const SetAccountValidPeriodMutationDocument = gql`
    mutation setAccountValidPeriodMutation($userID: ID!, $accountValidFrom: DateTime, $accountValidUntil: DateTime) {
  setAccountValidPeriod(
    input: {userID: $userID, accountValidFrom: $accountValidFrom, accountValidUntil: $accountValidUntil}
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
export type SetAccountValidPeriodMutationMutationFn = Apollo.MutationFunction<SetAccountValidPeriodMutationMutation, SetAccountValidPeriodMutationMutationVariables>;

/**
 * __useSetAccountValidPeriodMutationMutation__
 *
 * To run a mutation, you first call `useSetAccountValidPeriodMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetAccountValidPeriodMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setAccountValidPeriodMutationMutation, { data, loading, error }] = useSetAccountValidPeriodMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      accountValidFrom: // value for 'accountValidFrom'
 *      accountValidUntil: // value for 'accountValidUntil'
 *   },
 * });
 */
export function useSetAccountValidPeriodMutationMutation(baseOptions?: Apollo.MutationHookOptions<SetAccountValidPeriodMutationMutation, SetAccountValidPeriodMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetAccountValidPeriodMutationMutation, SetAccountValidPeriodMutationMutationVariables>(SetAccountValidPeriodMutationDocument, options);
      }
export type SetAccountValidPeriodMutationMutationHookResult = ReturnType<typeof useSetAccountValidPeriodMutationMutation>;
export type SetAccountValidPeriodMutationMutationResult = Apollo.MutationResult<SetAccountValidPeriodMutationMutation>;
export type SetAccountValidPeriodMutationMutationOptions = Apollo.BaseMutationOptions<SetAccountValidPeriodMutationMutation, SetAccountValidPeriodMutationMutationVariables>;