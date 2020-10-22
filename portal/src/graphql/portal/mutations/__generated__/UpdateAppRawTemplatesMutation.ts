/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppResourceUpdate } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppRawTemplatesMutation
// ====================================================

export interface UpdateAppRawTemplatesMutation_updateAppResources_app_resources {
  __typename: "AppResource";
  path: string;
  data: string | null;
}

export interface UpdateAppRawTemplatesMutation_updateAppResources_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  resources: UpdateAppRawTemplatesMutation_updateAppResources_app_resources[];
}

export interface UpdateAppRawTemplatesMutation_updateAppResources {
  __typename: "UpdateAppResourcesPayload";
  app: UpdateAppRawTemplatesMutation_updateAppResources_app;
}

export interface UpdateAppRawTemplatesMutation {
  /**
   * Update app resource files
   */
  updateAppResources: UpdateAppRawTemplatesMutation_updateAppResources;
}

export interface UpdateAppRawTemplatesMutationVariables {
  appID: string;
  updates: AppResourceUpdate[];
  paths: string[];
}
