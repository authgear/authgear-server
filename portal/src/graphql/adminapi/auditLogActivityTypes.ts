import { AuditLogActivityType } from "./globalTypes.generated";

export const ALL_ACTIVITY_TYPES = Object.values(AuditLogActivityType);

/**
 * Project-level activity that no prefix identifies. Every OAUTH_CLIENT_*
 * payload returns "" from UserID() (pkg/api/event/nonblocking/oauth_client_*.go),
 * so filing them under user activity is wrong twice over: the row shows no
 * user, and that tab's search is by user ID or email, which can never match
 * them. They describe the project's client set, and deleting such a client
 * already logs as ADMIN_API_MUTATION_DELETE_DYNAMIC_CLIENT_EXECUTED, so
 * filing them anywhere else would split one object's history across two
 * mutually exclusive tabs.
 */
const PROJECT_CLIENT_ACTIVITY_TYPES: AuditLogActivityType[] = [
  AuditLogActivityType.OauthClientRegistered,
  AuditLogActivityType.OauthClientRegistrationFailed,
  AuditLogActivityType.OauthClientResolved,
  AuditLogActivityType.OauthClientResolutionFailed,
];

export const PROJECT_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) =>
    activityType.startsWith("ADMIN_API") ||
    activityType.startsWith("PROJECT") ||
    PROJECT_CLIENT_ACTIVITY_TYPES.includes(activityType)
);

/** Activity types hidden from audit log UI (shown elsewhere in the portal). */
export const HIDDEN_ACTIVITY_TYPES = [
  AuditLogActivityType.FraudProtectionDecisionRecorded,
];

export const USER_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) =>
    !PROJECT_ACTIVITY_TYPES.includes(activityType) &&
    !HIDDEN_ACTIVITY_TYPES.includes(activityType)
);

// The values are the ?kind= URL param, so Admin keeps its wire value even
// though its tab now reads "Project Activities" -- renaming it would break
// existing bookmarks.
export enum AuditLogKind {
  User = "user",
  Admin = "admin",
}

export function isAuditLogKind(s: string): s is AuditLogKind {
  return Object.values(AuditLogKind).includes(s as AuditLogKind);
}
