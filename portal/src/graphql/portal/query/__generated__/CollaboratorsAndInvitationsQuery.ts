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

export interface CollaboratorsAndInvitationsQuery_node_App_collaborators_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  email: string | null;
}

export interface CollaboratorsAndInvitationsQuery_node_App_collaborators {
  __typename: "Collaborator";
  id: string;
  createdAt: GQL_DateTime;
  user: CollaboratorsAndInvitationsQuery_node_App_collaborators_user;
}

export interface CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations_invitedBy {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  email: string | null;
}

export interface CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations {
  __typename: "CollaboratorInvitation";
  id: string;
  createdAt: GQL_DateTime;
  expireAt: GQL_DateTime;
  invitedBy: CollaboratorsAndInvitationsQuery_node_App_collaboratorInvitations_invitedBy;
  inviteeEmail: string;
}

export interface CollaboratorsAndInvitationsQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
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
