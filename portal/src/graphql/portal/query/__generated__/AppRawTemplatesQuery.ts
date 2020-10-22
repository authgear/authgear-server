/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppRawTemplatesQuery
// ====================================================

export interface AppRawTemplatesQuery_node_User {
  __typename: "User";
}

export interface AppRawTemplatesQuery_node_App_resources {
  __typename: "AppResource";
  path: string;
  data: string | null;
}

export interface AppRawTemplatesQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  resources: AppRawTemplatesQuery_node_App_resources[];
}

export type AppRawTemplatesQuery_node = AppRawTemplatesQuery_node_User | AppRawTemplatesQuery_node_App;

export interface AppRawTemplatesQuery {
  /**
   * Fetches an object given its ID
   */
  node: AppRawTemplatesQuery_node | null;
}

export interface AppRawTemplatesQueryVariables {
  id: string;
  paths: string[];
}
