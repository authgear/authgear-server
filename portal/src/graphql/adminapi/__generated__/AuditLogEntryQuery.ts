/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AuditLogActivityType } from "./globalTypes";

// ====================================================
// GraphQL query operation: AuditLogEntryQuery
// ====================================================

export interface AuditLogEntryQuery_node_Authenticator {
  __typename: "Authenticator" | "Identity" | "Session" | "User";
}

export interface AuditLogEntryQuery_node_AuditLog_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
}

export interface AuditLogEntryQuery_node_AuditLog {
  __typename: "AuditLog";
  /**
   * The ID of an object
   */
  id: string;
  createdAt: GQL_DateTime;
  activityType: AuditLogActivityType;
  user: AuditLogEntryQuery_node_AuditLog_user | null;
  ipAddress: string | null;
  userAgent: string | null;
  clientID: string | null;
  data: GQL_AuditLogData | null;
}

export type AuditLogEntryQuery_node = AuditLogEntryQuery_node_Authenticator | AuditLogEntryQuery_node_AuditLog;

export interface AuditLogEntryQuery {
  /**
   * Fetches an object given its ID
   */
  node: AuditLogEntryQuery_node | null;
}

export interface AuditLogEntryQueryVariables {
  logID: string;
}
