/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteDomainMutation
// ====================================================

export interface DeleteDomainMutation_deleteDomain_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
}

export interface DeleteDomainMutation_deleteDomain {
  __typename: "DeleteDomainPayload";
  app: DeleteDomainMutation_deleteDomain_app;
}

export interface DeleteDomainMutation {
  /**
   * Delete domain of target app
   */
  deleteDomain: DeleteDomainMutation_deleteDomain;
}

export interface DeleteDomainMutationVariables {
  appID: string;
  domainID: string;
}
