/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteIdentityMutation
// ====================================================

export interface DeleteIdentityMutation_deleteIdentity_user_authenticators_edges_node {
  __typename: "Authenticator";
  /**
   * The ID of an object
   */
  id: string;
}

export interface DeleteIdentityMutation_deleteIdentity_user_authenticators_edges {
  __typename: "AuthenticatorEdge";
  /**
   * The item at the end of the edge
   */
  node: DeleteIdentityMutation_deleteIdentity_user_authenticators_edges_node | null;
}

export interface DeleteIdentityMutation_deleteIdentity_user_authenticators {
  __typename: "AuthenticatorConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (DeleteIdentityMutation_deleteIdentity_user_authenticators_edges | null)[] | null;
}

export interface DeleteIdentityMutation_deleteIdentity_user_identities_edges_node {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
}

export interface DeleteIdentityMutation_deleteIdentity_user_identities_edges {
  __typename: "IdentityEdge";
  /**
   * The item at the end of the edge
   */
  node: DeleteIdentityMutation_deleteIdentity_user_identities_edges_node | null;
}

export interface DeleteIdentityMutation_deleteIdentity_user_identities {
  __typename: "IdentityConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (DeleteIdentityMutation_deleteIdentity_user_identities_edges | null)[] | null;
}

export interface DeleteIdentityMutation_deleteIdentity_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  authenticators: DeleteIdentityMutation_deleteIdentity_user_authenticators | null;
  identities: DeleteIdentityMutation_deleteIdentity_user_identities | null;
}

export interface DeleteIdentityMutation_deleteIdentity {
  __typename: "DeleteIdentityPayload";
  user: DeleteIdentityMutation_deleteIdentity_user;
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
