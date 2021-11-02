/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { UserSortBy, SortDirection } from "./globalTypes";

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
  createdAt: GQL_DateTime;
  /**
   * The last login time of user
   */
  lastLoginAt: GQL_DateTime | null;
  isDisabled: boolean;
  standardAttributes: GQL_UserStandardAttributes;
}

export interface UsersListQuery_users_edges {
  __typename: "UserEdge";
  /**
   * The item at the end of the edge
   */
  node: UsersListQuery_users_edges_node | null;
}

export interface UsersListQuery_users {
  __typename: "UserConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (UsersListQuery_users_edges | null)[] | null;
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
  searchKeyword: string;
  pageSize: number;
  cursor?: string | null;
  sortBy?: UserSortBy | null;
  sortDirection?: SortDirection | null;
}
