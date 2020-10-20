/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppResourceUpdate } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppAndSecretConfigMutation
// ====================================================

export interface UpdateAppAndSecretConfigMutation_updateAppResources_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
  rawSecretConfig: GQL_SecretConfig;
}

export interface UpdateAppAndSecretConfigMutation_updateAppResources {
  __typename: "UpdateAppResourcesPayload";
  app: UpdateAppAndSecretConfigMutation_updateAppResources_app;
}

export interface UpdateAppAndSecretConfigMutation {
  /**
   * Update app resource files
   */
  updateAppResources: UpdateAppAndSecretConfigMutation_updateAppResources;
}

export interface UpdateAppAndSecretConfigMutationVariables {
  appID: string;
  updates: AppResourceUpdate[];
}
