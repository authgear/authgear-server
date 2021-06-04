/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AuditLogActivityType } from "./globalTypes";

// ====================================================
// GraphQL query operation: AuditLogListQuery
// ====================================================

export interface AuditLogListQuery_auditLogs_edges_node_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
}

export interface AuditLogListQuery_auditLogs_edges_node {
  __typename: "AuditLog";
  /**
   * The ID of an object
   */
  id: string;
  createdAt: GQL_DateTime;
  activityType: AuditLogActivityType;
  user: AuditLogListQuery_auditLogs_edges_node_user | null;
}

export interface AuditLogListQuery_auditLogs_edges {
  __typename: "AuditLogEdge";
  /**
   * The item at the end of the edge
   */
  node: AuditLogListQuery_auditLogs_edges_node | null;
}

export interface AuditLogListQuery_auditLogs {
  __typename: "AuditLogConnection";
  /**
   * Information to aid in pagination.
   */
  edges: (AuditLogListQuery_auditLogs_edges | null)[] | null;
  /**
   * Total number of nodes in the connection.
   */
  totalCount: number | null;
}

export interface AuditLogListQuery {
  /**
   * Audit logs
   */
  auditLogs: AuditLogListQuery_auditLogs | null;
}

export interface AuditLogListQueryVariables {
  pageSize: number;
  cursor?: string | null;
}
