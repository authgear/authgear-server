import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";

export type ActivityTypeCategoryId =
  | "user"
  | "email"
  | "phone"
  | "oauth-biometric"
  | "sms"
  | "whatsapp"
  | "security"
  | "usage"
  | "m2m"
  | "admin-api"
  | "project";

export const ACTIVITY_TYPE_CATEGORY_ORDER: ActivityTypeCategoryId[] = [
  "user",
  "email",
  "phone",
  "oauth-biometric",
  "sms",
  "whatsapp",
  "security",
  "usage",
  "m2m",
  "admin-api",
  "project",
];

export function getActivityTypeCategory(
  activityType: AuditLogActivityType
): ActivityTypeCategoryId {
  const activityTypeKey = activityType as string;
  if (activityTypeKey.startsWith("ADMIN_API_")) {
    return "admin-api";
  }
  if (activityTypeKey.startsWith("PROJECT_")) {
    return "project";
  }
  if (
    activityTypeKey.startsWith("USER_") ||
    activityTypeKey.startsWith("IDENTITY_USERNAME_") ||
    activityTypeKey.startsWith("AUTHENTICATION_")
  ) {
    return "user";
  }
  if (
    activityTypeKey.startsWith("IDENTITY_EMAIL_") ||
    activityTypeKey.startsWith("EMAIL_")
  ) {
    return "email";
  }
  if (activityTypeKey.startsWith("IDENTITY_PHONE_")) {
    return "phone";
  }
  if (
    activityTypeKey.startsWith("IDENTITY_OAUTH_") ||
    activityTypeKey.startsWith("IDENTITY_BIOMETRIC_")
  ) {
    return "oauth-biometric";
  }
  if (activityTypeKey.startsWith("SMS_")) {
    return "sms";
  }
  if (activityTypeKey.startsWith("WHATSAPP_")) {
    return "whatsapp";
  }
  if (
    activityTypeKey.startsWith("BOT_PROTECTION_") ||
    activityTypeKey.startsWith("RATE_LIMIT_")
  ) {
    return "security";
  }
  if (activityTypeKey.startsWith("USAGE_")) {
    return "usage";
  }
  if (activityTypeKey.startsWith("M2M_")) {
    return "m2m";
  }
  return "user";
}

export interface ActivityTypeCategoryGroup {
  id: ActivityTypeCategoryId;
  activityTypes: AuditLogActivityType[];
}

export function groupActivityTypesByCategory(
  activityTypes: AuditLogActivityType[]
): ActivityTypeCategoryGroup[] {
  const groups = new Map<ActivityTypeCategoryId, AuditLogActivityType[]>();
  for (const activityType of activityTypes) {
    const category = getActivityTypeCategory(activityType);
    const existing = groups.get(category) ?? [];
    existing.push(activityType);
    groups.set(category, existing);
  }

  return ACTIVITY_TYPE_CATEGORY_ORDER.filter((id) => groups.has(id)).map(
    (id) => ({
      id,
      activityTypes: groups.get(id)!,
    })
  );
}
