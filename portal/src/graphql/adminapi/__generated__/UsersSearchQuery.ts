/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: UsersSearchQuery
// ====================================================

export interface UsersSearchQuery_users_edges_node_identities_edges_node {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
  claims: GQL_IdentityClaims;
}

export interface UsersSearchQuery_users_edges_node_identities_edges {
  __typename: "IdentityEdge";
  /**
   * The item at the end of the edge
   */
  node: UsersSearchQuery_users_edges_node_identities_edges_node | null;
}

export interface UsersSearchQuery_users_edges_node_identities {
  __typename: "IdentityConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UsersSearchQuery_users_edges_node_identities_edges | null)[] | null;
}

export interface UsersSearchQuery_users_edges_node {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  /**
   * The creation time of entity
   */
  createdAt: GQL_DateTime;
  /**
   * The last login time of user
   */
  lastLoginAt: GQL_DateTime | null;
  isDisabled: boolean;
  identities: UsersSearchQuery_users_edges_node_identities | null;
}

export interface UsersSearchQuery_users_edges {
  __typename: "UserEdge";
  /**
   * The item at the end of the edge
   */
  node: UsersSearchQuery_users_edges_node | null;
}

export interface UsersSearchQuery_users {
  __typename: "UserConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UsersSearchQuery_users_edges | null)[] | null;
  /**
   * Total number of nodes in the connection.
   */
  totalCount: number | null;
}

export interface UsersSearchQuery {
  /**
   * Search users
   */
  users: UsersSearchQuery_users | null;
}

export interface UsersSearchQueryVariables {
  searchKeyword: string;
  pageSize: number;
  cursor?: string | null;
}
