import { AuditLogActivityType } from "./globalTypes.generated";

export const ALL_ACTIVITY_TYPES = Object.values(AuditLogActivityType);

export const ADMIN_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) =>
    activityType.startsWith("ADMIN_API") || activityType.startsWith("PROJECT")
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
