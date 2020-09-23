/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteIdentityMutation
// ====================================================

export interface DeleteIdentityMutation_deleteIdentity {
  __typename: "DeleteIdentityPayload";
  success: boolean;
}

export interface DeleteIdentityMutation {
  /**
   * Delete identity of user
   */
  deleteIdentity: DeleteIdentityMutation_deleteIdentity;
}

export interface DeleteIdentityMutationVariables {
  identityID: string;
}
