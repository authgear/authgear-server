/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: SetDisabledStatusMutation
// ====================================================

export interface SetDisabledStatusMutation_setDisabledStatus_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  isDisabled: boolean;
}

export interface SetDisabledStatusMutation_setDisabledStatus {
  __typename: "SetDisabledStatusPayload";
  user: SetDisabledStatusMutation_setDisabledStatus_user;
}

export interface SetDisabledStatusMutation {
  /**
   * Set disabled status of user
   */
  setDisabledStatus: SetDisabledStatusMutation_setDisabledStatus;
}

export interface SetDisabledStatusMutationVariables {
  userID: string;
  isDisabled: boolean;
}
