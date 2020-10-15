/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: CollaboratorsAndInvitationsQuery
// ====================================================

export interface CollaboratorsAndInvitationsQuery_node_User {
  __typename: "User";
}

export interface CollaboratorsAndInvitationsQuery_node_App_collaborators {
  __typename: "Collaborator";
  id: string;
  createdAt: GQL_DateTime;
  userID: string;
}

export interface CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations {
  __typename: "CollaboratorInvitation";
  id: string;
  createdAt: GQL_DateTime;
  expireAt: GQL_DateTime;
  invitedBy: string;
  inviteeEmail: string;
}

export interface CollaboratorsAndInvitationsQuery_node_App {
  __typename: "App";
  collaborators: CollaboratorsAndInvitationsQuery_node_App_collaborators[];
  collaboratorInvitations: CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations[];
}

export type CollaboratorsAndInvitationsQuery_node = CollaboratorsAndInvitationsQuery_node_User | CollaboratorsAndInvitationsQuery_node_App;

export interface CollaboratorsAndInvitationsQuery {
  /**
   * Fetches an object given its ID
   */
  node: CollaboratorsAndInvitationsQuery_node | null;
}

export interface CollaboratorsAndInvitationsQueryVariables {
  appID: string;
}
