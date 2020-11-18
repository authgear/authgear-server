/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AuthenticatorType, AuthenticatorKind, IdentityType, SessionType } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: UserQuery
// ====================================================

export interface UserQuery_node_Authenticator {
  __typename: "Authenticator" | "Identity" | "Session";
}

export interface UserQuery_node_User_authenticators_edges_node {
  __typename: "Authenticator";
  /**
   * The ID of an object
   */
  id: string;
  type: AuthenticatorType;
  kind: AuthenticatorKind;
  isDefault: boolean;
  claims: GQL_AuthenticatorClaims;
  /**
   * The creation time of entity
   */
  createdAt: GQL_DateTime;
  /**
   * The update time of entity
   */
  updatedAt: GQL_DateTime;
}

export interface UserQuery_node_User_authenticators_edges {
  __typename: "AuthenticatorEdge";
  /**
   * The item at the end of the edge
   */
  node: UserQuery_node_User_authenticators_edges_node | null;
}

export interface UserQuery_node_User_authenticators {
  __typename: "AuthenticatorConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UserQuery_node_User_authenticators_edges | null)[] | null;
}

export interface UserQuery_node_User_identities_edges_node {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
  type: IdentityType;
  claims: GQL_IdentityClaims;
  /**
   * The creation time of entity
   */
  createdAt: GQL_DateTime;
  /**
   * The update time of entity
   */
  updatedAt: GQL_DateTime;
}

export interface UserQuery_node_User_identities_edges {
  __typename: "IdentityEdge";
  /**
   * The item at the end of the edge
   */
  node: UserQuery_node_User_identities_edges_node | null;
}

export interface UserQuery_node_User_identities {
  __typename: "IdentityConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UserQuery_node_User_identities_edges | null)[] | null;
}

export interface UserQuery_node_User_verifiedClaims {
  __typename: "Claim";
  name: string;
  value: string;
}

export interface UserQuery_node_User_sessions_edges_node_userAgent {
  __typename: "UserAgent";
  name: string;
  version: string;
}

export interface UserQuery_node_User_sessions_edges_node {
  __typename: "Session";
  /**
   * The ID of an object
   */
  id: string;
  type: SessionType;
  lastAccessedAt: GQL_DateTime;
  lastAccessedByIP: string;
  userAgent: UserQuery_node_User_sessions_edges_node_userAgent;
}

export interface UserQuery_node_User_sessions_edges {
  __typename: "SessionEdge";
  /**
   * The item at the end of the edge
   */
  node: UserQuery_node_User_sessions_edges_node | null;
}

export interface UserQuery_node_User_sessions {
  __typename: "SessionConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UserQuery_node_User_sessions_edges | null)[] | null;
}

export interface UserQuery_node_User {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  authenticators: UserQuery_node_User_authenticators | null;
  identities: UserQuery_node_User_identities | null;
  verifiedClaims: UserQuery_node_User_verifiedClaims[];
  sessions: UserQuery_node_User_sessions | null;
  isDisabled: boolean;
  /**
   * The last login time of user
   */
  lastLoginAt: GQL_DateTime | null;
  /**
   * The creation time of entity
   */
  createdAt: GQL_DateTime;
  /**
   * The update time of entity
   */
  updatedAt: GQL_DateTime;
}

export type UserQuery_node = UserQuery_node_Authenticator | UserQuery_node_User;

export interface UserQuery {
  /**
   * Fetches an object given its ID
   */
  node: UserQuery_node | null;
}

export interface UserQueryVariables {
  userID: string;
}
