/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateUserMutation
// ====================================================

export interface UpdateUserMutation_updateUser_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  /**
   * The update time of entity
   */
  updatedAt: GQL_DateTime;
  standardAttributes: GQL_UserStandardAttributes;
}

export interface UpdateUserMutation_updateUser {
  __typename: "UpdateUserPayload";
  user: UpdateUserMutation_updateUser_user;
}

export interface UpdateUserMutation {
  /**
   * Update user
   */
  updateUser: UpdateUserMutation_updateUser;
}

export interface UpdateUserMutationVariables {
  userID: string;
  standardAttributes: GQL_UserStandardAttributes;
}
