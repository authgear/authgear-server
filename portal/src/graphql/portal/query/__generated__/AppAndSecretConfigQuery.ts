/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppAndSecretConfigQuery
// ====================================================

export interface AppAndSecretConfigQuery_node_User {
  __typename: "User";
}

export interface AppAndSecretConfigQuery_node_App_secretConfig_oauthClientSecrets {
  __typename: "OAuthClientSecret";
  alias: string;
  clientSecret: string;
}

export interface AppAndSecretConfigQuery_node_App_secretConfig_webhookSecret {
  __typename: "WebhookSecret";
  secret: string | null;
}

export interface AppAndSecretConfigQuery_node_App_secretConfig_adminAPISecrets {
  __typename: "AdminAPISecret";
  keyID: string;
  createdAt: GQL_DateTime | null;
  publicKeyPEM: string;
  privateKeyPEM: string | null;
}

export interface AppAndSecretConfigQuery_node_App_secretConfig {
  __typename: "SecretConfig";
  oauthClientSecrets: AppAndSecretConfigQuery_node_App_secretConfig_oauthClientSecrets[] | null;
  webhookSecret: AppAndSecretConfigQuery_node_App_secretConfig_webhookSecret | null;
  adminAPISecrets: AppAndSecretConfigQuery_node_App_secretConfig_adminAPISecrets[] | null;
}

export interface AppAndSecretConfigQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  effectiveAppConfig: GQL_AppConfig;
  rawAppConfig: GQL_AppConfig;
  secretConfig: AppAndSecretConfigQuery_node_App_secretConfig;
}

export type AppAndSecretConfigQuery_node = AppAndSecretConfigQuery_node_User | AppAndSecretConfigQuery_node_App;

export interface AppAndSecretConfigQuery {
  /**
   * Fetches an object given its ID
   */
  node: AppAndSecretConfigQuery_node | null;
}

export interface AppAndSecretConfigQueryVariables {
  id: string;
}
