/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: SkipAppTutorialProgressMutation
// ====================================================

export interface SkipAppTutorialProgressMutation_skipAppTutorialProgress_app_tutorialStatus {
  __typename: "TutorialStatus";
  data: GQL_TutorialStatusData;
}

export interface SkipAppTutorialProgressMutation_skipAppTutorialProgress_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  tutorialStatus: SkipAppTutorialProgressMutation_skipAppTutorialProgress_app_tutorialStatus;
}

export interface SkipAppTutorialProgressMutation_skipAppTutorialProgress {
  __typename: "SkipAppTutorialProgressPayload";
  app: SkipAppTutorialProgressMutation_skipAppTutorialProgress_app;
}

export interface SkipAppTutorialProgressMutation {
  /**
   * Skip a progress of the tutorial of the app
   */
  skipAppTutorialProgress: SkipAppTutorialProgressMutation_skipAppTutorialProgress;
}

export interface SkipAppTutorialProgressMutationVariables {
  appID: string;
  progress: string;
}
