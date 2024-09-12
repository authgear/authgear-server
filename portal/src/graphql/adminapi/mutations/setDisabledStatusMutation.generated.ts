import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetDisabledStatusMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  isDisabled: Types.Scalars['Boolean']['input'];
}>;


export type SetDisabledStatusMutationMutation = { __typename?: 'Mutation', setDisabledStatus: { __typename?: 'SetDisabledStatusPayload', user: { __typename?: 'User', id: string, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, isAnonymized: boolean, anonymizeAt?: any | null } } };


export const SetDisabledStatusMutationDocument = gql`
    mutation setDisabledStatusMutation($userID: ID!, $isDisabled: Boolean!) {
  setDisabledStatus(input: {userID: $userID, isDisabled: $isDisabled}) {
    user {
      id
      isDisabled
      disableReason
      isDeactivated
      deleteAt
      isAnonymized
      anonymizeAt
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