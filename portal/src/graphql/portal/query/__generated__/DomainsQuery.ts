/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: DomainsQuery
// ====================================================

export interface DomainsQuery_node_User {
  __typename: "User";
}

export interface DomainsQuery_node_App_domains {
  __typename: "Domain";
  id: string;
  createdAt: GQL_DateTime;
  apexDomain: string;
  domain: string;
  isCustom: boolean;
  isVerified: boolean;
  verificationDNSRecord: string;
}

export interface DomainsQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  domains: DomainsQuery_node_App_domains[];
}

export type DomainsQuery_node = DomainsQuery_node_User | DomainsQuery_node_App;

export interface DomainsQuery {
  /**
   * Fetches an object given its ID
   */
  node: DomainsQuery_node | null;
}

export interface DomainsQueryVariables {
  appID: string;
}
