/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: CreateAppMutation
// ====================================================

export interface CreateAppMutation_createApp_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
}

export interface CreateAppMutation_createApp {
  __typename: "CreateAppPayload";
  app: CreateAppMutation_createApp_app;
}

export interface CreateAppMutation {
  /**
   * Create new app
   */
  createApp: CreateAppMutation_createApp;
}

export interface CreateAppMutationVariables {
  appID: string;
}
