import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AnonymizeUserMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
}>;


export type AnonymizeUserMutationMutation = { __typename?: 'Mutation', anonymizeUser: { __typename?: 'AnonymizeUserPayload', anonymizedUserID: string } };


export const AnonymizeUserMutationDocument = gql`
    mutation anonymizeUserMutation($userID: ID!) {
  anonymizeUser(input: {userID: $userID}) {
    anonymizedUserID
  }
}
    `;
export type AnonymizeUserMutationMutationFn = Apollo.MutationFunction<AnonymizeUserMutationMutation, AnonymizeUserMutationMutationVariables>;

/**
 * __useAnonymizeUserMutationMutation__
 *
 * To run a mutation, you first call `useAnonymizeUserMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAnonymizeUserMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [anonymizeUserMutationMutation, { data, loading, error }] = useAnonymizeUserMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *   },
 * });
 */
export function useAnonymizeUserMutationMutation(baseOptions?: Apollo.MutationHookOptions<AnonymizeUserMutationMutation, AnonymizeUserMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AnonymizeUserMutationMutation, AnonymizeUserMutationMutationVariables>(AnonymizeUserMutationDocument, options);
      }
export type AnonymizeUserMutationMutationHookResult = ReturnType<typeof useAnonymizeUserMutationMutation>;
export type AnonymizeUserMutationMutationResult = Apollo.MutationResult<AnonymizeUserMutationMutation>;
export type AnonymizeUserMutationMutationOptions = Apollo.BaseMutationOptions<AnonymizeUserMutationMutation, AnonymizeUserMutationMutationVariables>;