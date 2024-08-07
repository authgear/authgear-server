import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import { AuthenticatorFragmentFragmentDoc } from '../query/userQuery.generated';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SetPasswordExpiredMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  expired: Types.Scalars['Boolean']['input'];
}>;


export type SetPasswordExpiredMutation = { __typename?: 'Mutation', setPasswordExpired: { __typename?: 'SetPasswordExpiredPayload', user: { __typename?: 'User', id: string, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string, type: Types.AuthenticatorType, kind: Types.AuthenticatorKind, isDefault: boolean, claims: any, createdAt: any, updatedAt: any, expireAfter?: any | null } | null } | null> | null } | null } } };


export const SetPasswordExpiredDocument = gql`
    mutation setPasswordExpired($userID: ID!, $expired: Boolean!) {
  setPasswordExpired(input: {userID: $userID, expired: $expired}) {
    user {
      id
      authenticators {
        edges {
          node {
            ...AuthenticatorFragment
          }
        }
      }
    }
  }
}
    ${AuthenticatorFragmentFragmentDoc}`;
export type SetPasswordExpiredMutationFn = Apollo.MutationFunction<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>;

/**
 * __useSetPasswordExpiredMutation__
 *
 * To run a mutation, you first call `useSetPasswordExpiredMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetPasswordExpiredMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setPasswordExpiredMutation, { data, loading, error }] = useSetPasswordExpiredMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      expired: // value for 'expired'
 *   },
 * });
 */
export function useSetPasswordExpiredMutation(baseOptions?: Apollo.MutationHookOptions<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>(SetPasswordExpiredDocument, options);
      }
export type SetPasswordExpiredMutationHookResult = ReturnType<typeof useSetPasswordExpiredMutation>;
export type SetPasswordExpiredMutationResult = Apollo.MutationResult<SetPasswordExpiredMutation>;
export type SetPasswordExpiredMutationOptions = Apollo.BaseMutationOptions<SetPasswordExpiredMutation, SetPasswordExpiredMutationVariables>;