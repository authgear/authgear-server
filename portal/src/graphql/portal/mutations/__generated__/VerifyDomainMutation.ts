/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: VerifyDomainMutation
// ====================================================

export interface VerifyDomainMutation_verifyDomain_app_domains {
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

export interface VerifyDomainMutation_verifyDomain_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  domains: VerifyDomainMutation_verifyDomain_app_domains[];
}

export interface VerifyDomainMutation_verifyDomain_domain {
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

export interface VerifyDomainMutation_verifyDomain {
  __typename: "VerifyDomainPayload";
  app: VerifyDomainMutation_verifyDomain_app;
  domain: VerifyDomainMutation_verifyDomain_domain;
}

export interface VerifyDomainMutation {
  /**
   * Request verification of a domain of target app
   */
  verifyDomain: VerifyDomainMutation_verifyDomain;
}

export interface VerifyDomainMutationVariables {
  appID: string;
  domainID: string;
}
