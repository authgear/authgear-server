/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppConfigFile } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppAndEmailSmsTemplatesConfigMutation
// ====================================================

export interface UpdateAppAndEmailSmsTemplatesConfigMutation_updateAppConfig {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
}

export interface UpdateAppAndEmailSmsTemplatesConfigMutation {
  /**
   * Update app configuration files
   */
  updateAppConfig: UpdateAppAndEmailSmsTemplatesConfigMutation_updateAppConfig;
}

export interface UpdateAppAndEmailSmsTemplatesConfigMutationVariables {
  appID: string;
  updateFiles: AppConfigFile[];
}
