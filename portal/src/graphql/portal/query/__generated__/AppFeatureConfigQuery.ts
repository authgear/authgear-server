/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppFeatureConfigQuery
// ====================================================

export interface AppFeatureConfigQuery_node_User {
  __typename: "User";
}

export interface AppFeatureConfigQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  effectiveFeatureConfig: GQL_FeatureConfig;
}

export type AppFeatureConfigQuery_node = AppFeatureConfigQuery_node_User | AppFeatureConfigQuery_node_App;

export interface AppFeatureConfigQuery {
  /**
   * Fetches an object given its ID
   */
  node: AppFeatureConfigQuery_node | null;
}

export interface AppFeatureConfigQueryVariables {
  id: string;
}
