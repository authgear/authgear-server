import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateAuthenticatorMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  definition: Types.AuthenticatorDefinition;
}>;


export type CreateAuthenticatorMutationMutation = { __typename?: 'Mutation', createAuthenticator: { __typename?: 'CreateAuthenticatorPayload', authenticator: { __typename?: 'Authenticator', id: string } } };


export const CreateAuthenticatorMutationDocument = gql`
    mutation createAuthenticatorMutation($userID: ID!, $definition: AuthenticatorDefinition!) {
  createAuthenticator(input: {userID: $userID, definition: $definition}) {
    authenticator {
      id
    }
  }
}
    `;
export type CreateAuthenticatorMutationMutationFn = Apollo.MutationFunction<CreateAuthenticatorMutationMutation, CreateAuthenticatorMutationMutationVariables>;

/**
 * __useCreateAuthenticatorMutationMutation__
 *
 * To run a mutation, you first call `useCreateAuthenticatorMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateAuthenticatorMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createAuthenticatorMutationMutation, { data, loading, error }] = useCreateAuthenticatorMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      definition: // value for 'definition'
 *   },
 * });
 */
export function useCreateAuthenticatorMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateAuthenticatorMutationMutation, CreateAuthenticatorMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateAuthenticatorMutationMutation, CreateAuthenticatorMutationMutationVariables>(CreateAuthenticatorMutationDocument, options);
      }
export type CreateAuthenticatorMutationMutationHookResult = ReturnType<typeof useCreateAuthenticatorMutationMutation>;
export type CreateAuthenticatorMutationMutationResult = Apollo.MutationResult<CreateAuthenticatorMutationMutation>;
export type CreateAuthenticatorMutationMutationOptions = Apollo.BaseMutationOptions<CreateAuthenticatorMutationMutation, CreateAuthenticatorMutationMutationVariables>;