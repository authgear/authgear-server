/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: AcceptCollaboratorInvitationMutation
// ====================================================

export interface AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation_app_collaborators {
  __typename: "Collaborator";
  id: string;
  createdAt: GQL_DateTime;
  userID: string;
}

export interface AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation_app_collaboratorInvitations {
  __typename: "CollaboratorInvitation";
  id: string;
  createdAt: GQL_DateTime;
  expireAt: GQL_DateTime;
  invitedBy: string;
  inviteeEmail: string;
}

export interface AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  collaborators: AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation_app_collaborators[];
  collaboratorInvitations: AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation_app_collaboratorInvitations[];
}

export interface AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation {
  __typename: "AcceptCollaboratorInvitationPayload";
  app: AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation_app;
}

export interface AcceptCollaboratorInvitationMutation {
  /**
   * Accept collaborator invitation to the target app.
   */
  acceptCollaboratorInvitation: AcceptCollaboratorInvitationMutation_acceptCollaboratorInvitation;
}

export interface AcceptCollaboratorInvitationMutationVariables {
  code: string;
}
