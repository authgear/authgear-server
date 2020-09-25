/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteAuthenticatorMutation
// ====================================================

export interface DeleteAuthenticatorMutation_deleteAuthenticator {
  __typename: "DeleteAuthenticatorPayload";
  success: boolean;
}

export interface DeleteAuthenticatorMutation {
  /**
   * Delete authenticator of user
   */
  deleteAuthenticator: DeleteAuthenticatorMutation_deleteAuthenticator;
}

export interface DeleteAuthenticatorMutationVariables {
  authenticatorID: string;
}
