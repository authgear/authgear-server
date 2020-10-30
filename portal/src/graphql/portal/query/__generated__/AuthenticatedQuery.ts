/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AuthenticatedQuery
// ====================================================

export interface AuthenticatedQuery_viewer {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
}

export interface AuthenticatedQuery {
  /**
   * The current viewer
   */
  viewer: AuthenticatedQuery_viewer | null;
}
