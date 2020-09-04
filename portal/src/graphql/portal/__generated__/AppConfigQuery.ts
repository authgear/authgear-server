/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppConfigQuery
// ====================================================

export interface AppConfigQuery_node_User {
  __typename: "User";
}

export interface AppConfigQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  appConfig: GQL_JSONObject;
}

export type AppConfigQuery_node = AppConfigQuery_node_User | AppConfigQuery_node_App;

export interface AppConfigQuery {
  /**
   * Fetches an object given its ID
   */
  node: AppConfigQuery_node | null;
}

export interface AppConfigQueryVariables {
  id: string;
}
