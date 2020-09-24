/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: ResetPasswordMutation
// ====================================================

export interface ResetPasswordMutation_resetPassword_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
}

export interface ResetPasswordMutation_resetPassword {
  __typename: "ResetPasswordPayload";
  user: ResetPasswordMutation_resetPassword_user;
}

export interface ResetPasswordMutation {
  /**
   * Reset password of user
   */
  resetPassword: ResetPasswordMutation_resetPassword;
}

export interface ResetPasswordMutationVariables {
  userID: string;
  password: string;
}
