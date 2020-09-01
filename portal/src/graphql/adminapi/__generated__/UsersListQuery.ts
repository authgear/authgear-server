/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: UsersListQuery
// ====================================================

export interface UsersListQuery_users_edges_node {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  /**
   * The creation time of entity
   */
  createdAt: any;
}

export interface UsersListQuery_users_edges {
  __typename: "UserEdge";
  /**
   * The item at the end of the edge
   */
  node: UsersListQuery_users_edges_node | null;
}

export interface UsersListQuery_users_pageInfo {
  __typename: "PageInfo";
  /**
   * When paginating forwards, are there more items?
   */
  hasNextPage: boolean;
  /**
   * When paginating forwards, the cursor to continue.
   */
  endCursor: string | null;
}

export interface UsersListQuery_users {
  __typename: "UserConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UsersListQuery_users_edges | null)[] | null;
  /**
   * Information to aid in pagination.
   */
  pageInfo: UsersListQuery_users_pageInfo;
  /**
   * Total number of nodes in the connection.
   */
  totalCount: number | null;
}

export interface UsersListQuery {
  /**
   * All users
   */
  users: UsersListQuery_users | null;
}

export interface UsersListQueryVariables {
  pageSize: number;
  cursor?: string | null;
}
