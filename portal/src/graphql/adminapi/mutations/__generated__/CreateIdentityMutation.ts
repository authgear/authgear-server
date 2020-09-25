/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { IdentityDefinition } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: CreateIdentityMutation
// ====================================================

export interface CreateIdentityMutation_createIdentity_identity {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
  claims: GQL_IdentityClaims;
}

export interface CreateIdentityMutation_createIdentity {
  __typename: "CreateIdentityPayload";
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
