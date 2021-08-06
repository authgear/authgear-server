/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AuthenticatedForInvitationQuery
// ====================================================

export interface AuthenticatedForInvitationQuery_viewer {
  __typename: "User";
  email: string | null;
}

export interface AuthenticatedForInvitationQuery_checkCollaboratorInvitation {
  __typename: "CheckCollaboratorInvitationPayload";
  isCodeValid: boolean;
  isInvitee: boolean;
  appID: string;
}

export interface AuthenticatedForInvitationQuery {
  /**
   * The current viewer
   */
  viewer: AuthenticatedForInvitationQuery_viewer | null;
  /**
   * Check whether the viewer can accept the collaboration invitation
   */
  checkCollaboratorInvitation: AuthenticatedForInvitationQuery_checkCollaboratorInvitation;
}

export interface AuthenticatedForInvitationQueryVariables {
  code: string;
}
