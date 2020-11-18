/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteUserMutation
// ====================================================

export interface DeleteUserMutation_deleteUser {
  __typename: "DeleteUserPayload";
  deletedUserID: string;
}

export interface DeleteUserMutation {
  /**
   * Delete specified user
   */
  deleteUser: DeleteUserMutation_deleteUser;
}

export interface DeleteUserMutationVariables {
  userID: string;
}
