import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type MarkPasswordAsExpiredMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  isExpired: Types.Scalars['Boolean']['input'];
}>;


export type MarkPasswordAsExpiredMutation = { __typename?: 'Mutation', markPasswordAsExpired: { __typename?: 'MarkPasswordAsExpiredPayload', user: { __typename?: 'User', id: string } } };


export const MarkPasswordAsExpiredDocument = gql`
    mutation markPasswordAsExpired($userID: ID!, $isExpired: Boolean!) {
  markPasswordAsExpired(input: {userID: $userID, isExpired: $isExpired}) {
    user {
      id
    }
  }
}
    `;
export type MarkPasswordAsExpiredMutationFn = Apollo.MutationFunction<MarkPasswordAsExpiredMutation, MarkPasswordAsExpiredMutationVariables>;

/**
 * __useMarkPasswordAsExpiredMutation__
 *
 * To run a mutation, you first call `useMarkPasswordAsExpiredMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useMarkPasswordAsExpiredMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [markPasswordAsExpiredMutation, { data, loading, error }] = useMarkPasswordAsExpiredMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      isExpired: // value for 'isExpired'
 *   },
 * });
 */
export function useMarkPasswordAsExpiredMutation(baseOptions?: Apollo.MutationHookOptions<MarkPasswordAsExpiredMutation, MarkPasswordAsExpiredMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<MarkPasswordAsExpiredMutation, MarkPasswordAsExpiredMutationVariables>(MarkPasswordAsExpiredDocument, options);
      }
export type MarkPasswordAsExpiredMutationHookResult = ReturnType<typeof useMarkPasswordAsExpiredMutation>;
export type MarkPasswordAsExpiredMutationResult = Apollo.MutationResult<MarkPasswordAsExpiredMutation>;
export type MarkPasswordAsExpiredMutationOptions = Apollo.BaseMutationOptions<MarkPasswordAsExpiredMutation, MarkPasswordAsExpiredMutationVariables>;