/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppListQuery
// ====================================================

export interface AppListQuery_apps_edges_node {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  effectiveAppConfig: GQL_AppConfig;
}

export interface AppListQuery_apps_edges {
  __typename: "AppEdge";
  /**
   * The item at the end of the edge
   */
  node: AppListQuery_apps_edges_node | null;
}

export interface AppListQuery_apps {
  __typename: "AppConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (AppListQuery_apps_edges | null)[] | null;
}

export interface AppListQuery {
  /**
   * All apps accessible by the viewer
   */
  apps: AppListQuery_apps | null;
}
