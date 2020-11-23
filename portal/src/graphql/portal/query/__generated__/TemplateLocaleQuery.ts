/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: TemplateLocaleQuery
// ====================================================

export interface TemplateLocaleQuery_node_User {
  __typename: "User";
}

export interface TemplateLocaleQuery_node_App_resourceLocales {
  __typename: "AppResource";
  path: string;
  languageTag: string | null;
}

export interface TemplateLocaleQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  resourceLocales: TemplateLocaleQuery_node_App_resourceLocales[];
}

export type TemplateLocaleQuery_node = TemplateLocaleQuery_node_User | TemplateLocaleQuery_node_App;

export interface TemplateLocaleQuery {
  /**
   * Fetches an object given its ID
   */
  node: TemplateLocaleQuery_node | null;
}

export interface TemplateLocaleQueryVariables {
  id: string;
}
