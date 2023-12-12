import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RevokeAllSessionsMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type RevokeAllSessionsMutationMutation = { __typename?: 'Mutation', revokeAllSessions: { __typename?: 'RevokeAllSessionsPayload', user: { __typename?: 'User', id: string, sessions?: { __typename?: 'SessionConnection', edges?: Array<{ __typename?: 'SessionEdge', node?: { __typename?: 'Session', id: string } | null } | null> | null } | null } } };


export const RevokeAllSessionsMutationDocument = gql`
    mutation revokeAllSessionsMutation($userID: ID!) {
  revokeAllSessions(input: {userID: $userID}) {
    user {
      id
      sessions {
        edges {
          node {
            id
          }
        }
      }
    }
  }
}
    `;
export type RevokeAllSessionsMutationMutationFn = Apollo.MutationFunction<RevokeAllSessionsMutationMutation, RevokeAllSessionsMutationMutationVariables>;

/**
 * __useRevokeAllSessionsMutationMutation__
 *
 * To run a mutation, you first call `useRevokeAllSessionsMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRevokeAllSessionsMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [revokeAllSessionsMutationMutation, { data, loading, error }] = useRevokeAllSessionsMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useRevokeAllSessionsMutationMutation(baseOptions?: Apollo.MutationHookOptions<RevokeAllSessionsMutationMutation, RevokeAllSessionsMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RevokeAllSessionsMutationMutation, RevokeAllSessionsMutationMutationVariables>(RevokeAllSessionsMutationDocument, options);
      }
export type RevokeAllSessionsMutationMutationHookResult = ReturnType<typeof useRevokeAllSessionsMutationMutation>;
export type RevokeAllSessionsMutationMutationResult = Apollo.MutationResult<RevokeAllSessionsMutationMutation>;
export type RevokeAllSessionsMutationMutationOptions = Apollo.BaseMutationOptions<RevokeAllSessionsMutationMutation, RevokeAllSessionsMutationMutationVariables>;