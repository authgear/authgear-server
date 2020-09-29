/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { IdentityDefinition, IdentityType } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: CreateIdentityMutation
// ====================================================

export interface CreateIdentityMutation_createIdentity_user_authenticators_edges_node {
  __typename: "Authenticator";
  /**
   * The ID of an object
   */
  id: string;
}

export interface CreateIdentityMutation_createIdentity_user_authenticators_edges {
  __typename: "AuthenticatorEdge";
  /**
   * The item at the end of the edge
   */
  node: CreateIdentityMutation_createIdentity_user_authenticators_edges_node | null;
}

export interface CreateIdentityMutation_createIdentity_user_authenticators {
  __typename: "AuthenticatorConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (CreateIdentityMutation_createIdentity_user_authenticators_edges | null)[] | null;
}

export interface CreateIdentityMutation_createIdentity_user_identities_edges_node {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
}

export interface CreateIdentityMutation_createIdentity_user_identities_edges {
  __typename: "IdentityEdge";
  /**
   * The item at the end of the edge
   */
  node: CreateIdentityMutation_createIdentity_user_identities_edges_node | null;
}

export interface CreateIdentityMutation_createIdentity_user_identities {
  __typename: "IdentityConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (CreateIdentityMutation_createIdentity_user_identities_edges | null)[] | null;
}

export interface CreateIdentityMutation_createIdentity_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  authenticators: CreateIdentityMutation_createIdentity_user_authenticators | null;
  identities: CreateIdentityMutation_createIdentity_user_identities | null;
}

export interface CreateIdentityMutation_createIdentity_identity {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
  type: IdentityType;
  claims: GQL_IdentityClaims;
  /**
   * The creation time of entity
   */
  createdAt: GQL_DateTime;
  /**
   * The update time of entity
   */
  updatedAt: GQL_DateTime;
}

export interface CreateIdentityMutation_createIdentity {
  __typename: "CreateIdentityPayload";
  user: CreateIdentityMutation_createIdentity_user;
  identity: CreateIdentityMutation_createIdentity_identity;
}

export interface CreateIdentityMutation {
  /**
   * Create new identity for user
   */
  createIdentity: CreateIdentityMutation_createIdentity;
}

export interface CreateIdentityMutationVariables {
  userID: string;
  definition: IdentityDefinition;
  password?: string | null;
}
