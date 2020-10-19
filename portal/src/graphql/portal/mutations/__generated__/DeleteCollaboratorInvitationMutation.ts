/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteCollaboratorInvitationMutation
// ====================================================

export interface DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation_app_collaboratorInvitations_invitedBy {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  email: string | null;
}

export interface DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation_app_collaboratorInvitations {
  __typename: "CollaboratorInvitation";
  id: string;
  createdAt: GQL_DateTime;
  expireAt: GQL_DateTime;
  invitedBy: DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation_app_collaboratorInvitations_invitedBy;
  inviteeEmail: string;
}

export interface DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  collaboratorInvitations: DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation_app_collaboratorInvitations[];
}

export interface DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation {
  __typename: "DeleteCollaboratorInvitationPayload";
  app: DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation_app;
}

export interface DeleteCollaboratorInvitationMutation {
  /**
   * Delete collaborator invitation of target app.
   */
  deleteCollaboratorInvitation: DeleteCollaboratorInvitationMutation_deleteCollaboratorInvitation;
}

export interface DeleteCollaboratorInvitationMutationVariables {
  collaboratorInvitationID: string;
}
