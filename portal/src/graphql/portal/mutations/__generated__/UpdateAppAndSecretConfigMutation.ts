/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { SecretConfigInput } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppAndSecretConfigMutation
// ====================================================

export interface UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_oauthClientSecrets {
  __typename: "OAuthClientSecret";
  alias: string;
  clientSecret: string;
}

export interface UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_webhookSecret {
  __typename: "WebhookSecret";
  secret: string | null;
}

export interface UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_adminAPISecrets {
  __typename: "AdminAPISecret";
  keyID: string;
  createdAt: GQL_DateTime | null;
  publicKeyPEM: string;
  privateKeyPEM: string | null;
}

export interface UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_smtpSecret {
  __typename: "SMTPSecret";
  host: string;
  port: number;
  username: string;
  password: string | null;
}

export interface UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig {
  __typename: "SecretConfig";
  oauthClientSecrets: UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_oauthClientSecrets[] | null;
  webhookSecret: UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_webhookSecret | null;
  adminAPISecrets: UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_adminAPISecrets[] | null;
  smtpSecret: UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig_smtpSecret | null;
}

export interface UpdateAppAndSecretConfigMutation_updateApp_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  effectiveAppConfig: GQL_AppConfig;
  rawAppConfig: GQL_AppConfig;
  secretConfig: UpdateAppAndSecretConfigMutation_updateApp_app_secretConfig;
}

export interface UpdateAppAndSecretConfigMutation_updateApp {
  __typename: "UpdateAppPayload";
  app: UpdateAppAndSecretConfigMutation_updateApp_app;
}

export interface UpdateAppAndSecretConfigMutation {
  /**
   * Update app
   */
  updateApp: UpdateAppAndSecretConfigMutation_updateApp;
}

export interface UpdateAppAndSecretConfigMutationVariables {
  appID: string;
  appConfig: GQL_AppConfig;
  secretConfig?: SecretConfigInput | null;
}
