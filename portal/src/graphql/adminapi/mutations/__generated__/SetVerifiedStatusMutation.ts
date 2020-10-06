/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: SetVerifiedStatusMutation
// ====================================================

export interface SetVerifiedStatusMutation_setVerifiedStatus_user_identities_edges_node {
  __typename: "Identity";
  /**
   * The ID of an object
   */
  id: string;
  claims: GQL_IdentityClaims;
}

export interface SetVerifiedStatusMutation_setVerifiedStatus_user_identities_edges {
  __typename: "IdentityEdge";
  /**
   * The item at the end of the edge
   */
  node: SetVerifiedStatusMutation_setVerifiedStatus_user_identities_edges_node | null;
}

export interface SetVerifiedStatusMutation_setVerifiedStatus_user_identities {
  __typename: "IdentityConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (SetVerifiedStatusMutation_setVerifiedStatus_user_identities_edges | null)[] | null;
}

export interface SetVerifiedStatusMutation_setVerifiedStatus_user_verifiedClaims {
  __typename: "Claim";
  name: string;
  value: string;
}

export interface SetVerifiedStatusMutation_setVerifiedStatus_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  identities: SetVerifiedStatusMutation_setVerifiedStatus_user_identities | null;
  verifiedClaims: SetVerifiedStatusMutation_setVerifiedStatus_user_verifiedClaims[];
}

export interface SetVerifiedStatusMutation_setVerifiedStatus {
  __typename: "SetVerifiedStatusPayload";
  user: SetVerifiedStatusMutation_setVerifiedStatus_user;
}

export interface SetVerifiedStatusMutation {
  /**
   * Set verified status of a claim of user
   */
  setVerifiedStatus: SetVerifiedStatusMutation_setVerifiedStatus;
}

export interface SetVerifiedStatusMutationVariables {
  userID: string;
  claimName: string;
  claimValue: string;
  isVerified: boolean;
}
