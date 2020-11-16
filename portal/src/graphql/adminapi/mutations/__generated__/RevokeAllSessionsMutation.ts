/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: RevokeAllSessionsMutation
// ====================================================

export interface RevokeAllSessionsMutation_revokeAllSessions_user_sessions_edges_node {
  __typename: "Session";
  /**
   * The ID of an object
   */
  id: string;
}

export interface RevokeAllSessionsMutation_revokeAllSessions_user_sessions_edges {
  __typename: "SessionEdge";
  /**
   * The item at the end of the edge
   */
  node: RevokeAllSessionsMutation_revokeAllSessions_user_sessions_edges_node | null;
}

export interface RevokeAllSessionsMutation_revokeAllSessions_user_sessions {
  __typename: "SessionConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (RevokeAllSessionsMutation_revokeAllSessions_user_sessions_edges | null)[] | null;
}

export interface RevokeAllSessionsMutation_revokeAllSessions_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  sessions: RevokeAllSessionsMutation_revokeAllSessions_user_sessions | null;
}

export interface RevokeAllSessionsMutation_revokeAllSessions {
  __typename: "RevokeAllSessionsPayload";
  user: RevokeAllSessionsMutation_revokeAllSessions_user;
}

export interface RevokeAllSessionsMutation {
  /**
   * Revoke all sessions of user
   */
  revokeAllSessions: RevokeAllSessionsMutation_revokeAllSessions;
}

export interface RevokeAllSessionsMutationVariables {
  userID: string;
}
