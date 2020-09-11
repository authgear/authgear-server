/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppConfigFile } from "./../../../../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppConfigMutation
// ====================================================

export interface UpdateAppConfigMutation_updateAppConfig {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
}

export interface UpdateAppConfigMutation {
  /**
   * Update app configuration files
   */
  updateAppConfig: UpdateAppConfigMutation_updateAppConfig;
}

export interface UpdateAppConfigMutationVariables {
  appID: string;
  updateFile: AppConfigFile;
}
