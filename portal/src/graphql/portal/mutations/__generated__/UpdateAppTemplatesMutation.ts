/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AppResourceUpdate } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAppTemplatesMutation
// ====================================================

export interface UpdateAppTemplatesMutation_updateAppResources_app_resources {
  __typename: "AppResource";
  path: string;
  languageTag: string | null;
  data: string | null;
  effectiveData: string | null;
}

export interface UpdateAppTemplatesMutation_updateAppResources_app_resourceLocales {
  __typename: "AppResource";
  path: string;
  languageTag: string | null;
}

export interface UpdateAppTemplatesMutation_updateAppResources_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  resources: UpdateAppTemplatesMutation_updateAppResources_app_resources[];
  resourceLocales: UpdateAppTemplatesMutation_updateAppResources_app_resourceLocales[];
}

export interface UpdateAppTemplatesMutation_updateAppResources {
  __typename: "UpdateAppResourcesPayload";
  app: UpdateAppTemplatesMutation_updateAppResources_app;
}

export interface UpdateAppTemplatesMutation {
  /**
   * Update app resource files
   */
  updateAppResources: UpdateAppTemplatesMutation_updateAppResources;
}

export interface UpdateAppTemplatesMutationVariables {
  appID: string;
  updates: AppResourceUpdate[];
  paths: string[];
}
