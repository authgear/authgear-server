/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: CreateCollaboratorInvitationMutation
// ====================================================

export interface CreateCollaboratorInvitationMutation_createCollaboratorInvitation_collaboratorInvitation {
  __typename: "CollaboratorInvitation";
  id: string;
  createdAt: GQL_DateTime;
  expireAt: GQL_DateTime;
  invitedBy: string;
  inviteeEmail: string;
}

export interface CreateCollaboratorInvitationMutation_createCollaboratorInvitation {
  __typename: "CreateCollaboratorInvitationPayload";
  collaboratorInvitation: CreateCollaboratorInvitationMutation_createCollaboratorInvitation_collaboratorInvitation;
}

export interface CreateCollaboratorInvitationMutation {
  /**
   * Invite a collaborator to the target app.
   */
  createCollaboratorInvitation: CreateCollaboratorInvitationMutation_createCollaboratorInvitation;
}

export interface CreateCollaboratorInvitationMutationVariables {
  appID: string;
  email: string;
}
