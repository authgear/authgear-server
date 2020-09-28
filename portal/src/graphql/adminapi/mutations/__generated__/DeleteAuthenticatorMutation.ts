/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteAuthenticatorMutation
// ====================================================

export interface DeleteAuthenticatorMutation_deleteAuthenticator_user_authenticators_edges_node {
  __typename: "Authenticator";
  /**
   * The ID of an object
   */
  id: string;
}

export interface DeleteAuthenticatorMutation_deleteAuthenticator_user_authenticators_edges {
  __typename: "AuthenticatorEdge";
  /**
   * The item at the end of the edge
   */
  node: DeleteAuthenticatorMutation_deleteAuthenticator_user_authenticators_edges_node | null;
}

export interface DeleteAuthenticatorMutation_deleteAuthenticator_user_authenticators {
  __typename: "AuthenticatorConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (DeleteAuthenticatorMutation_deleteAuthenticator_user_authenticators_edges | null)[] | null;
}

export interface DeleteAuthenticatorMutation_deleteAuthenticator_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  authenticators: DeleteAuthenticatorMutation_deleteAuthenticator_user_authenticators | null;
}

export interface DeleteAuthenticatorMutation_deleteAuthenticator {
  __typename: "DeleteAuthenticatorPayload";
  user: DeleteAuthenticatorMutation_deleteAuthenticator_user;
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
