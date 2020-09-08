/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: UserDetailsScreenQuery
// ====================================================

export interface UserDetailsScreenQuery_node_Authenticator {
  __typename: "Authenticator" | "Identity";
}

export interface UserDetailsScreenQuery_node_User_authenticators_edges_node {
  __typename: "Authenticator";
  /**
   * The ID of an object
   */
  id: string;
  type: string;
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

export interface UserDetailsScreenQuery_node_User_authenticators_edges {
  __typename: "AuthenticatorEdge";
  /**
   * The item at the end of the edge
   */
  node: UserDetailsScreenQuery_node_User_authenticators_edges_node | null;
}

export interface UserDetailsScreenQuery_node_User_authenticators {
  __typename: "AuthenticatorConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UserDetailsScreenQuery_node_User_authenticators_edges | null)[] | null;
}

export interface UserDetailsScreenQuery_node_User_identities_edges_node {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
  type: string;
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

export interface UserDetailsScreenQuery_node_User_identities_edges {
  __typename: "IdentityEdge";
  /**
   * The item at the end of the edge
   */
  node: UserDetailsScreenQuery_node_User_identities_edges_node | null;
}

export interface UserDetailsScreenQuery_node_User_identities {
  __typename: "IdentityConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UserDetailsScreenQuery_node_User_identities_edges | null)[] | null;
}

export interface UserDetailsScreenQuery_node_User {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  authenticators: UserDetailsScreenQuery_node_User_authenticators | null;
  identities: UserDetailsScreenQuery_node_User_identities | null;
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

export type UserDetailsScreenQuery_node = UserDetailsScreenQuery_node_Authenticator | UserDetailsScreenQuery_node_User;

export interface UserDetailsScreenQuery {
  /**
   * Fetches an object given its ID
   */
  node: UserDetailsScreenQuery_node | null;
}

export interface UserDetailsScreenQueryVariables {
  userID: string;
}
