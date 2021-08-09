/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ViewerQuery
// ====================================================

export interface ViewerQuery_viewer {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  email: string | null;
}

export interface ViewerQuery {
  /**
   * The current viewer
   */
  viewer: ViewerQuery_viewer | null;
}
