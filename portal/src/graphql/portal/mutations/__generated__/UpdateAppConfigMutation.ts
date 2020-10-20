/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppResourceUpdate } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppConfigMutation
// ====================================================

export interface UpdateAppConfigMutation_updateAppResources_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
}

export interface UpdateAppConfigMutation_updateAppResources {
  __typename: "UpdateAppResourcesPayload";
  app: UpdateAppConfigMutation_updateAppResources_app;
}

export interface UpdateAppConfigMutation {
  /**
   * Update app resource files
   */
  updateAppResources: UpdateAppConfigMutation_updateAppResources;
}

export interface UpdateAppConfigMutationVariables {
  appID: string;
  updates: AppResourceUpdate[];
}
