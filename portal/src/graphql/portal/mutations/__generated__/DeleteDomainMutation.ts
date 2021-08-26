/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteDomainMutation
// ====================================================

export interface DeleteDomainMutation_deleteDomain_app_domains {
  __typename: "Domain";
  id: string;
  createdAt: GQL_DateTime;
  domain: string;
  cookieDomain: string;
  apexDomain: string;
  isCustom: boolean;
  isVerified: boolean;
  verificationDNSRecord: string;
}

export interface DeleteDomainMutation_deleteDomain_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  domains: DeleteDomainMutation_deleteDomain_app_domains[];
  rawAppConfig: GQL_AppConfig;
  effectiveAppConfig: GQL_AppConfig;
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
