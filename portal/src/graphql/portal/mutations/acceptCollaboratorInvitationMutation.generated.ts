import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AcceptCollaboratorInvitationMutationMutationVariables = Types.Exact<{
  code: Types.Scalars['String']['input'];
}>;


export type AcceptCollaboratorInvitationMutationMutation = { __typename?: 'Mutation', acceptCollaboratorInvitation: { __typename?: 'AcceptCollaboratorInvitationPayload', app: { __typename?: 'App', id: string, collaborators: Array<{ __typename?: 'Collaborator', id: string, createdAt: any, user: { __typename?: 'User', id: string, email?: string | null } }>, collaboratorInvitations: Array<{ __typename?: 'CollaboratorInvitation', id: string, createdAt: any, expireAt: any, inviteeEmail: string, invitedBy: { __typename?: 'User', id: string, email?: string | null } }> } } };


export const AcceptCollaboratorInvitationMutationDocument = gql`
    mutation acceptCollaboratorInvitationMutation($code: String!) {
  acceptCollaboratorInvitation(input: {code: $code}) {
    app {
      id
      collaborators {
        id
        createdAt
        user {
          id
          email
        }
      }
      collaboratorInvitations {
        id
        createdAt
        expireAt
        invitedBy {
          id
          email
        }
        inviteeEmail
      }
    }
  }
}
    `;
export type AcceptCollaboratorInvitationMutationMutationFn = Apollo.MutationFunction<AcceptCollaboratorInvitationMutationMutation, AcceptCollaboratorInvitationMutationMutationVariables>;

/**
 * __useAcceptCollaboratorInvitationMutationMutation__
 *
 * To run a mutation, you first call `useAcceptCollaboratorInvitationMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAcceptCollaboratorInvitationMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [acceptCollaboratorInvitationMutationMutation, { data, loading, error }] = useAcceptCollaboratorInvitationMutationMutation({
 *   variables: {
 *      code: // value for 'code'
 *   },
 * });
 */
export function useAcceptCollaboratorInvitationMutationMutation(baseOptions?: Apollo.MutationHookOptions<AcceptCollaboratorInvitationMutationMutation, AcceptCollaboratorInvitationMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AcceptCollaboratorInvitationMutationMutation, AcceptCollaboratorInvitationMutationMutationVariables>(AcceptCollaboratorInvitationMutationDocument, options);
      }
export type AcceptCollaboratorInvitationMutationMutationHookResult = ReturnType<typeof useAcceptCollaboratorInvitationMutationMutation>;
export type AcceptCollaboratorInvitationMutationMutationResult = Apollo.MutationResult<AcceptCollaboratorInvitationMutationMutation>;
export type AcceptCollaboratorInvitationMutationMutationOptions = Apollo.BaseMutationOptions<AcceptCollaboratorInvitationMutationMutation, AcceptCollaboratorInvitationMutationMutationVariables>;