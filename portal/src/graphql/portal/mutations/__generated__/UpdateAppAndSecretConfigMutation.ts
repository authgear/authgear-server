/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppConfigFile } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppAndSecretConfigMutation
// ====================================================

export interface UpdateAppAndSecretConfigMutation_updateAppConfig {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
  rawSecretConfig: GQL_SecretConfig;
}

export interface UpdateAppAndSecretConfigMutation {
  /**
   * Update app configuration files
   */
  updateAppConfig: UpdateAppAndSecretConfigMutation_updateAppConfig;
}

export interface UpdateAppAndSecretConfigMutationVariables {
  appID: string;
  appConfigFile: AppConfigFile;
  secretConfigFile: AppConfigFile;
}
