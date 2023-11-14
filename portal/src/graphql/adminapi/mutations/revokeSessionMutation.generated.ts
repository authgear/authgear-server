import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type RevokeSessionMutationMutationVariables = Types.Exact<{
  sessionID: Types.Scalars['ID']['input'];
}>;


export type RevokeSessionMutationMutation = { __typename?: 'Mutation', revokeSession: { __typename?: 'RevokeSessionPayload', user: { __typename?: 'User', id: string, sessions?: { __typename?: 'SessionConnection', edges?: Array<{ __typename?: 'SessionEdge', node?: { __typename?: 'Session', id: string } | null } | null> | null } | null } } };


export const RevokeSessionMutationDocument = gql`
    mutation revokeSessionMutation($sessionID: ID!) {
  revokeSession(input: {sessionID: $sessionID}) {
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
export type RevokeSessionMutationMutationFn = Apollo.MutationFunction<RevokeSessionMutationMutation, RevokeSessionMutationMutationVariables>;

/**
 * __useRevokeSessionMutationMutation__
 *
 * To run a mutation, you first call `useRevokeSessionMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRevokeSessionMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [revokeSessionMutationMutation, { data, loading, error }] = useRevokeSessionMutationMutation({
 *   variables: {
 *      sessionID: // value for 'sessionID'
 *   },
 * });
 */
export function useRevokeSessionMutationMutation(baseOptions?: Apollo.MutationHookOptions<RevokeSessionMutationMutation, RevokeSessionMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RevokeSessionMutationMutation, RevokeSessionMutationMutationVariables>(RevokeSessionMutationDocument, options);
      }
export type RevokeSessionMutationMutationHookResult = ReturnType<typeof useRevokeSessionMutationMutation>;
export type RevokeSessionMutationMutationResult = Apollo.MutationResult<RevokeSessionMutationMutation>;
export type RevokeSessionMutationMutationOptions = Apollo.BaseMutationOptions<RevokeSessionMutationMutation, RevokeSessionMutationMutationVariables>;