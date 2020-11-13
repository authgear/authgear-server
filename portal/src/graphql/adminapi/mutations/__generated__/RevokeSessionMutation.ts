/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: RevokeSessionMutation
// ====================================================

export interface RevokeSessionMutation_revokeSession_user_sessions_edges_node {
  __typename: "Session";
  /**
   * The ID of an object
   */
  id: string;
}

export interface RevokeSessionMutation_revokeSession_user_sessions_edges {
  __typename: "SessionEdge";
  /**
   * The item at the end of the edge
   */
  node: RevokeSessionMutation_revokeSession_user_sessions_edges_node | null;
}

export interface RevokeSessionMutation_revokeSession_user_sessions {
  __typename: "SessionConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (RevokeSessionMutation_revokeSession_user_sessions_edges | null)[] | null;
}

export interface RevokeSessionMutation_revokeSession_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  sessions: RevokeSessionMutation_revokeSession_user_sessions | null;
}

export interface RevokeSessionMutation_revokeSession {
  __typename: "RevokeSessionPayload";
  user: RevokeSessionMutation_revokeSession_user;
}

export interface RevokeSessionMutation {
  /**
   * Revoke session of user
   */
  revokeSession: RevokeSessionMutation_revokeSession;
}

export interface RevokeSessionMutationVariables {
  sessionID: string;
}
