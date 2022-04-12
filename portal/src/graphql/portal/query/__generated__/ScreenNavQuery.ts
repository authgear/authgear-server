/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ScreenNavQuery
// ====================================================

export interface ScreenNavQuery_node_User {
  __typename: "User";
}

export interface ScreenNavQuery_node_App_tutorialStatus {
  __typename: "TutorialStatus";
  appID: string;
  data: GQL_TutorialStatusData;
}

export interface ScreenNavQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  effectiveFeatureConfig: GQL_FeatureConfig;
  planName: string;
  tutorialStatus: ScreenNavQuery_node_App_tutorialStatus;
}

export type ScreenNavQuery_node = ScreenNavQuery_node_User | ScreenNavQuery_node_App;

export interface ScreenNavQuery {
  /**
   * Fetches an object given its ID
   */
  node: ScreenNavQuery_node | null;
}

export interface ScreenNavQueryVariables {
  id: string;
}
