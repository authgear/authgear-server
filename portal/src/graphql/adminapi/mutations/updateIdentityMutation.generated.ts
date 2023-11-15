import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateIdentityMutationMutationVariables = Types.Exact<{
  userID: Types.Scalars['ID']['input'];
  identityID: Types.Scalars['ID']['input'];
  definition: Types.IdentityDefinition;
}>;


export type UpdateIdentityMutationMutation = { __typename?: 'Mutation', updateIdentity: { __typename?: 'UpdateIdentityPayload', user: { __typename?: 'User', id: string, standardAttributes: any, customAttributes: any, web3: any, endUserAccountID?: string | null, updatedAt: any, authenticators?: { __typename?: 'AuthenticatorConnection', edges?: Array<{ __typename?: 'AuthenticatorEdge', node?: { __typename?: 'Authenticator', id: string, type: Types.AuthenticatorType, kind: Types.AuthenticatorKind, isDefault: boolean, claims: any, createdAt: any, updatedAt: any } | null } | null> | null } | null, identities?: { __typename?: 'IdentityConnection', edges?: Array<{ __typename?: 'IdentityEdge', node?: { __typename?: 'Identity', id: string, type: Types.IdentityType, claims: any, createdAt: any, updatedAt: any } | null } | null> | null } | null, verifiedClaims: Array<{ __typename?: 'Claim', name: string, value: string }> }, identity: { __typename?: 'Identity', id: string, type: Types.IdentityType, claims: any, createdAt: any, updatedAt: any } } };


export const UpdateIdentityMutationDocument = gql`
    mutation updateIdentityMutation($userID: ID!, $identityID: ID!, $definition: IdentityDefinition!) {
  updateIdentity(
    input: {userID: $userID, identityID: $identityID, definition: $definition}
  ) {
    user {
      id
      authenticators {
        edges {
          node {
            id
            type
            kind
            isDefault
            claims
            createdAt
            updatedAt
          }
        }
      }
      identities {
        edges {
          node {
            id
            type
            claims
            createdAt
            updatedAt
          }
        }
      }
      verifiedClaims {
        name
        value
      }
      standardAttributes
      customAttributes
      web3
      endUserAccountID
      updatedAt
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
export type UpdateIdentityMutationMutationFn = Apollo.MutationFunction<UpdateIdentityMutationMutation, UpdateIdentityMutationMutationVariables>;

/**
 * __useUpdateIdentityMutationMutation__
 *
 * To run a mutation, you first call `useUpdateIdentityMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateIdentityMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateIdentityMutationMutation, { data, loading, error }] = useUpdateIdentityMutationMutation({
 *   variables: {
 *      userID: // value for 'userID'
 *      identityID: // value for 'identityID'
 *      definition: // value for 'definition'
 *   },
 * });
 */
export function useUpdateIdentityMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateIdentityMutationMutation, UpdateIdentityMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateIdentityMutationMutation, UpdateIdentityMutationMutationVariables>(UpdateIdentityMutationDocument, options);
      }
export type UpdateIdentityMutationMutationHookResult = ReturnType<typeof useUpdateIdentityMutationMutation>;
export type UpdateIdentityMutationMutationResult = Apollo.MutationResult<UpdateIdentityMutationMutation>;
export type UpdateIdentityMutationMutationOptions = Apollo.BaseMutationOptions<UpdateIdentityMutationMutation, UpdateIdentityMutationMutationVariables>;