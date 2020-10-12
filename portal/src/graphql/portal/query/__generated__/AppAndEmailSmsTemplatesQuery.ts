/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppAndEmailSmsTemplatesQuery
// ====================================================

export interface AppAndEmailSmsTemplatesQuery_node_User {
  __typename: "User";
}

export interface AppAndEmailSmsTemplatesQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
  emailHtml: string;
  emailMjml: string;
  emailText: string;
  smsText: string;
}

export type AppAndEmailSmsTemplatesQuery_node = AppAndEmailSmsTemplatesQuery_node_User | AppAndEmailSmsTemplatesQuery_node_App;

export interface AppAndEmailSmsTemplatesQuery {
  /**
   * Fetches an object given its ID
   */
  node: AppAndEmailSmsTemplatesQuery_node | null;
}

export interface AppAndEmailSmsTemplatesQueryVariables {
  id: string;
  emailHtmlPath: string;
  emailMjmlPath: string;
  emailTextPath: string;
  smsTextPath: string;
}
