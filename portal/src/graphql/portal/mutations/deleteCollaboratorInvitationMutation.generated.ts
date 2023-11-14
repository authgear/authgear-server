import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteCollaboratorInvitationMutationMutationVariables = Types.Exact<{
  collaboratorInvitationID: Types.Scalars['String']['input'];
}>;


export type DeleteCollaboratorInvitationMutationMutation = { __typename?: 'Mutation', deleteCollaboratorInvitation: { __typename?: 'DeleteCollaboratorInvitationPayload', app: { __typename?: 'App', id: string, collaboratorInvitations: Array<{ __typename?: 'CollaboratorInvitation', id: string, createdAt: any, expireAt: any, inviteeEmail: string, invitedBy: { __typename?: 'User', id: string, email?: string | null } }> } } };


export const DeleteCollaboratorInvitationMutationDocument = gql`
    mutation deleteCollaboratorInvitationMutation($collaboratorInvitationID: String!) {
  deleteCollaboratorInvitation(
    input: {collaboratorInvitationID: $collaboratorInvitationID}
  ) {
    app {
      id
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
export type DeleteCollaboratorInvitationMutationMutationFn = Apollo.MutationFunction<DeleteCollaboratorInvitationMutationMutation, DeleteCollaboratorInvitationMutationMutationVariables>;

/**
 * __useDeleteCollaboratorInvitationMutationMutation__
 *
 * To run a mutation, you first call `useDeleteCollaboratorInvitationMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteCollaboratorInvitationMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteCollaboratorInvitationMutationMutation, { data, loading, error }] = useDeleteCollaboratorInvitationMutationMutation({
 *   variables: {
 *      collaboratorInvitationID: // value for 'collaboratorInvitationID'
 *   },
 * });
 */
export function useDeleteCollaboratorInvitationMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteCollaboratorInvitationMutationMutation, DeleteCollaboratorInvitationMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteCollaboratorInvitationMutationMutation, DeleteCollaboratorInvitationMutationMutationVariables>(DeleteCollaboratorInvitationMutationDocument, options);
      }
export type DeleteCollaboratorInvitationMutationMutationHookResult = ReturnType<typeof useDeleteCollaboratorInvitationMutationMutation>;
export type DeleteCollaboratorInvitationMutationMutationResult = Apollo.MutationResult<DeleteCollaboratorInvitationMutationMutation>;
export type DeleteCollaboratorInvitationMutationMutationOptions = Apollo.BaseMutationOptions<DeleteCollaboratorInvitationMutationMutation, DeleteCollaboratorInvitationMutationMutationVariables>;