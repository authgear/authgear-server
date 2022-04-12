/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: SkipAppTutorialMutation
// ====================================================

export interface SkipAppTutorialMutation_skipAppTutorial_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
}

export interface SkipAppTutorialMutation_skipAppTutorial {
  __typename: "SkipAppTutorialPayload";
  app: SkipAppTutorialMutation_skipAppTutorial_app;
}

export interface SkipAppTutorialMutation {
  /**
   * Skip the tutorial of the app
   */
  skipAppTutorial: SkipAppTutorialMutation_skipAppTutorial;
}

export interface SkipAppTutorialMutationVariables {
  appID: string;
}
