import { AuditLogActivityType } from "./globalTypes.generated";

export const ALL_ACTIVITY_TYPES = Object.values(AuditLogActivityType);

/**
 * Project-level activity that no prefix identifies. Every OAUTH_CLIENT_*
 * payload returns "" from UserID() (pkg/api/event/nonblocking/oauth_client_*.go),
 * so filing them under user activity is wrong twice over: the row shows no
 * user, and that tab's search is by user ID or email, which can never match
 * them. They describe the project's client set, so they belong with the
 * admin/project types.
 */
const PROJECT_CLIENT_ACTIVITY_TYPES: AuditLogActivityType[] = [
  AuditLogActivityType.OauthClientRegistered,
  AuditLogActivityType.OauthClientRegistrationFailed,
  AuditLogActivityType.OauthClientResolved,
  AuditLogActivityType.OauthClientResolutionFailed,
];

export const ADMIN_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
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
    !ADMIN_ACTIVITY_TYPES.includes(activityType) &&
    !HIDDEN_ACTIVITY_TYPES.includes(activityType)
);

export enum AuditLogKind {
  User = "user",
  Admin = "admin",
}

export function isAuditLogKind(s: string): s is AuditLogKind {
  return Object.values(AuditLogKind).includes(s as AuditLogKind);
}
