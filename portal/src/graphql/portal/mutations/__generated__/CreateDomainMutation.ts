/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: CreateDomainMutation
// ====================================================

export interface CreateDomainMutation_createDomain_app_domains {
  __typename: "Domain";
  id: string;
  createdAt: GQL_DateTime;
  domain: string;
  apexDomain: string;
  isCustom: boolean;
  isVerified: boolean;
  verificationDNSRecord: string;
}

export interface CreateDomainMutation_createDomain_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  domains: CreateDomainMutation_createDomain_app_domains[];
}

export interface CreateDomainMutation_createDomain_domain {
  __typename: "Domain";
  id: string;
  createdAt: GQL_DateTime;
  domain: string;
  apexDomain: string;
  isCustom: boolean;
  isVerified: boolean;
  verificationDNSRecord: string;
}

export interface CreateDomainMutation_createDomain {
  __typename: "CreateDomainPayload";
  app: CreateDomainMutation_createDomain_app;
  domain: CreateDomainMutation_createDomain_domain;
}

export interface CreateDomainMutation {
  /**
   * Create domain for target app
   */
  createDomain: CreateDomainMutation_createDomain;
}

export interface CreateDomainMutationVariables {
  appID: string;
  domain: string;
}
