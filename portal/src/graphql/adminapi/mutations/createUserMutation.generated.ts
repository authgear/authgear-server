import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateUserMutationMutationVariables = Types.Exact<{
  identityDefinition: Types.IdentityDefinitionLoginId;
  password?: Types.InputMaybe<Types.Scalars['String']['input']>;
  sendPassword?: Types.InputMaybe<Types.Scalars['Boolean']['input']>;
  setPasswordExpired?: Types.InputMaybe<Types.Scalars['Boolean']['input']>;
}>;


export type CreateUserMutationMutation = { __typename?: 'Mutation', createUser: { __typename?: 'CreateUserPayload', user: { __typename?: 'User', id: string } } };


export const CreateUserMutationDocument = gql`
    mutation createUserMutation($identityDefinition: IdentityDefinitionLoginID!, $password: String, $sendPassword: Boolean, $setPasswordExpired: Boolean) {
  createUser(
    input: {definition: {loginID: $identityDefinition}, password: $password, sendPassword: $sendPassword, setPasswordExpired: $setPasswordExpired}
  ) {
    user {
      id
    }
  }
}
    `;
export type CreateUserMutationMutationFn = Apollo.MutationFunction<CreateUserMutationMutation, CreateUserMutationMutationVariables>;

/**
 * __useCreateUserMutationMutation__
 *
 * To run a mutation, you first call `useCreateUserMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateUserMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createUserMutationMutation, { data, loading, error }] = useCreateUserMutationMutation({
 *   variables: {
 *      identityDefinition: // value for 'identityDefinition'
 *      password: // value for 'password'
 *      sendPassword: // value for 'sendPassword'
 *      setPasswordExpired: // value for 'setPasswordExpired'
 *   },
 * });
 */
export function useCreateUserMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateUserMutationMutation, CreateUserMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateUserMutationMutation, CreateUserMutationMutationVariables>(CreateUserMutationDocument, options);
      }
export type CreateUserMutationMutationHookResult = ReturnType<typeof useCreateUserMutationMutation>;
export type CreateUserMutationMutationResult = Apollo.MutationResult<CreateUserMutationMutation>;
export type CreateUserMutationMutationOptions = Apollo.BaseMutationOptions<CreateUserMutationMutation, CreateUserMutationMutationVariables>;