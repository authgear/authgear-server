/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppsScreenQuery
// ====================================================

export interface AppsScreenQuery_apps_edges_node {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  effectiveAppConfig: GQL_AppConfig;
}

export interface AppsScreenQuery_apps_edges {
  __typename: "AppEdge";
  /**
   * The item at the end of the edge
   */
  node: AppsScreenQuery_apps_edges_node | null;
}

export interface AppsScreenQuery_apps {
  __typename: "AppConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (AppsScreenQuery_apps_edges | null)[] | null;
}

export interface AppsScreenQuery {
  /**
   * All apps accessible by the viewer
   */
  apps: AppsScreenQuery_apps | null;
}
