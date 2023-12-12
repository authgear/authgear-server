import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateIdentityMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  definition: Types.IdentityDefinition;
  password?: Types.InputMaybe<Types.Scalars['String']['input']>;
}>;


export type CreateIdentityMutationMutation = { __typename?: 'Mutation', createIdentity: { __typename?: 'CreateIdentityPayload', user: { __typename?: 'User', id: string, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string } | null } | null> | null } | null, identities?: { __typename?: 'IdentityConnection', edges?: Array<{ __typename?: 'IdentityEdge', node?: { __typename?: 'Identity', id: string } | null } | null> | null } | null }, identity: { __typename?: 'Identity', id: string, type: Types.IdentityType, claims: any, createdAt: any, updatedAt: any } } };


export const CreateIdentityMutationDocument = gql`
    mutation createIdentityMutation($userID: ID!, $definition: IdentityDefinition!, $password: String) {
  createIdentity(
    input: {userID: $userID, definition: $definition, password: $password}
  ) {
    user {
      id
      authenticators {
        edges {
          node {
            id
          }
        }
      }
      identities {
        edges {
          node {
            id
          }
        }
      }
    }
    identity {
      id
      type
      claims
      createdAt
      updatedAt
    }
  }
}
    `;
export type CreateIdentityMutationMutationFn = Apollo.MutationFunction<CreateIdentityMutationMutation, CreateIdentityMutationMutationVariables>;

/**
 * __useCreateIdentityMutationMutation__
 *
 * To run a mutation, you first call `useCreateIdentityMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateIdentityMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createIdentityMutationMutation, { data, loading, error }] = useCreateIdentityMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      definition: // value for 'definition'
 *      password: // value for 'password'
 *   },
 * });
 */
export function useCreateIdentityMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateIdentityMutationMutation, CreateIdentityMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateIdentityMutationMutation, CreateIdentityMutationMutationVariables>(CreateIdentityMutationDocument, options);
      }
export type CreateIdentityMutationMutationHookResult = ReturnType<typeof useCreateIdentityMutationMutation>;
export type CreateIdentityMutationMutationResult = Apollo.MutationResult<CreateIdentityMutationMutation>;
export type CreateIdentityMutationMutationOptions = Apollo.BaseMutationOptions<CreateIdentityMutationMutation, CreateIdentityMutationMutationVariables>;