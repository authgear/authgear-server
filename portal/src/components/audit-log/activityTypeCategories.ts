import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";

export type ActivityTypeCategoryGroupId =
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

export type ActivityTypeSubcategoryId =
  | "user-account"
  | "user-authentication"
  | "user-username"
  | "email-delivery"
  | "email-identity"
  | "phone-identity"
  | "oauth-identity"
  | "biometric-identity"
  | "sms"
  | "whatsapp"
  | "security"
  | "usage"
  | "m2m"
  | "admin-api-user"
  | "admin-api-identity"
  | "admin-api-session"
  | "admin-api-password"
  | "admin-api-group-role"
  | "admin-api-oauth"
  | "project-app"
  | "project-billing"
  | "project-collaborator"
  | "project-domain";

/** @deprecated Use ActivityTypeSubcategoryId instead. Kept for call-site compatibility. */
export type ActivityTypeCategoryId = ActivityTypeSubcategoryId;

export const ACTIVITY_TYPE_CATEGORY_GROUP_ORDER: ActivityTypeCategoryGroupId[] =
  [
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

const GROUP_SUBCATEGORY_ORDER: Record<
  ActivityTypeCategoryGroupId,
  ActivityTypeSubcategoryId[]
> = {
  user: ["user-account", "user-authentication", "user-username"],
  email: ["email-delivery", "email-identity"],
  phone: ["phone-identity"],
  "oauth-biometric": ["oauth-identity", "biometric-identity"],
  sms: ["sms"],
  whatsapp: ["whatsapp"],
  security: ["security"],
  usage: ["usage"],
  m2m: ["m2m"],
  "admin-api": [
    "admin-api-user",
    "admin-api-identity",
    "admin-api-session",
    "admin-api-password",
    "admin-api-group-role",
    "admin-api-oauth",
  ],
  project: [
    "project-app",
    "project-billing",
    "project-collaborator",
    "project-domain",
  ],
};

const SUBCATEGORY_TO_GROUP: Record<
  ActivityTypeSubcategoryId,
  ActivityTypeCategoryGroupId
> = {
  "user-account": "user",
  "user-authentication": "user",
  "user-username": "user",
  "email-delivery": "email",
  "email-identity": "email",
  "phone-identity": "phone",
  "oauth-identity": "oauth-biometric",
  "biometric-identity": "oauth-biometric",
  sms: "sms",
  whatsapp: "whatsapp",
  security: "security",
  usage: "usage",
  m2m: "m2m",
  "admin-api-user": "admin-api",
  "admin-api-identity": "admin-api",
  "admin-api-session": "admin-api",
  "admin-api-password": "admin-api",
  "admin-api-group-role": "admin-api",
  "admin-api-oauth": "admin-api",
  "project-app": "project",
  "project-billing": "project",
  "project-collaborator": "project",
  "project-domain": "project",
};

function getAdminApiActivityTypeSubcategory(
  activityTypeKey: string
): ActivityTypeSubcategoryId {
  const mutation = activityTypeKey
    .replace(/^ADMIN_API_MUTATION_/, "")
    .replace(/_EXECUTED$/, "");

  if (
    mutation.includes("RESOURCE") ||
    mutation.includes("SCOPE") ||
    mutation.includes("CLIENTID") ||
    mutation.includes("SCOPES")
  ) {
    return "admin-api-oauth";
  }

  if (mutation.includes("GROUP") || mutation.includes("ROLE")) {
    return "admin-api-group-role";
  }

  if (mutation.includes("SESSION") || mutation.includes("AUTHORIZATION")) {
    return "admin-api-session";
  }

  if (mutation.includes("IDENTITY") || mutation.includes("AUTHENTICATOR")) {
    return "admin-api-identity";
  }

  if (mutation.includes("PASSWORD") || mutation.includes("OOB_OTP")) {
    if (mutation === "SET_PASSWORD_EXPIRED") {
      return "admin-api-user";
    }
    return "admin-api-password";
  }

  return "admin-api-user";
}

function getProjectActivityTypeSubcategory(
  activityTypeKey: string
): ActivityTypeSubcategoryId {
  if (activityTypeKey.startsWith("PROJECT_APP_")) {
    return "project-app";
  }
  if (activityTypeKey.startsWith("PROJECT_BILLING_")) {
    return "project-billing";
  }
  if (activityTypeKey.startsWith("PROJECT_COLLABORATOR_")) {
    return "project-collaborator";
  }
  if (activityTypeKey.startsWith("PROJECT_DOMAIN_")) {
    return "project-domain";
  }
  return "project-app";
}

export function getActivityTypeSubcategory(
  activityType: AuditLogActivityType
): ActivityTypeSubcategoryId {
  const activityTypeKey = activityType as string;
  if (activityTypeKey.startsWith("ADMIN_API_")) {
    return getAdminApiActivityTypeSubcategory(activityTypeKey);
  }
  if (activityTypeKey.startsWith("PROJECT_")) {
    return getProjectActivityTypeSubcategory(activityTypeKey);
  }
  if (activityTypeKey.startsWith("AUTHENTICATION_")) {
    return "user-authentication";
  }
  if (activityTypeKey.startsWith("IDENTITY_USERNAME_")) {
    return "user-username";
  }
  if (activityTypeKey.startsWith("USER_")) {
    return "user-account";
  }
  if (activityTypeKey.startsWith("IDENTITY_EMAIL_")) {
    return "email-identity";
  }
  if (activityTypeKey.startsWith("EMAIL_")) {
    return "email-delivery";
  }
  if (activityTypeKey.startsWith("IDENTITY_PHONE_")) {
    return "phone-identity";
  }
  if (activityTypeKey.startsWith("IDENTITY_OAUTH_")) {
    return "oauth-identity";
  }
  if (activityTypeKey.startsWith("IDENTITY_BIOMETRIC_")) {
    return "biometric-identity";
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
  return "user-account";
}

/** @deprecated Use getActivityTypeSubcategory instead. */
export function getActivityTypeCategory(
  activityType: AuditLogActivityType
): ActivityTypeSubcategoryId {
  return getActivityTypeSubcategory(activityType);
}

export function getActivityTypeCategoryGroup(
  subcategoryId: ActivityTypeSubcategoryId
): ActivityTypeCategoryGroupId {
  return SUBCATEGORY_TO_GROUP[subcategoryId];
}

export function hasMultipleSubcategories(
  groupId: ActivityTypeCategoryGroupId
): boolean {
  return GROUP_SUBCATEGORY_ORDER[groupId].length > 1;
}

export interface ActivityTypeSubcategoryGroup {
  id: ActivityTypeSubcategoryId;
  activityTypes: AuditLogActivityType[];
}

export interface ActivityTypeCategoryGroupSection {
  id: ActivityTypeCategoryGroupId;
  subcategories: ActivityTypeSubcategoryGroup[];
}

export function groupActivityTypesByHierarchy(
  activityTypes: AuditLogActivityType[]
): ActivityTypeCategoryGroupSection[] {
  const subcategoryMap = new Map<
    ActivityTypeSubcategoryId,
    AuditLogActivityType[]
  >();
  for (const activityType of activityTypes) {
    const subcategory = getActivityTypeSubcategory(activityType);
    const existing = subcategoryMap.get(subcategory) ?? [];
    existing.push(activityType);
    subcategoryMap.set(subcategory, existing);
  }

  return ACTIVITY_TYPE_CATEGORY_GROUP_ORDER.map((groupId) => {
    const subcategories = GROUP_SUBCATEGORY_ORDER[groupId]
      .filter((subcategoryId) => subcategoryMap.has(subcategoryId))
      .map((subcategoryId) => ({
        id: subcategoryId,
        activityTypes: subcategoryMap.get(subcategoryId)!,
      }));

    if (subcategories.length === 0) {
      return null;
    }

    return {
      id: groupId,
      subcategories,
    };
  }).filter((section): section is ActivityTypeCategoryGroupSection => {
    return section != null;
  });
}

/** @deprecated Use groupActivityTypesByHierarchy instead. */
export interface ActivityTypeCategoryGroup {
  id: ActivityTypeSubcategoryId;
  activityTypes: AuditLogActivityType[];
}

/** @deprecated Use groupActivityTypesByHierarchy instead. */
export function groupActivityTypesByCategory(
  activityTypes: AuditLogActivityType[]
): ActivityTypeCategoryGroup[] {
  const groups = new Map<ActivityTypeSubcategoryId, AuditLogActivityType[]>();
  for (const activityType of activityTypes) {
    const category = getActivityTypeSubcategory(activityType);
    const existing = groups.get(category) ?? [];
    existing.push(activityType);
    groups.set(category, existing);
  }

  const orderedSubcategories = ACTIVITY_TYPE_CATEGORY_GROUP_ORDER.flatMap(
    (groupId) => GROUP_SUBCATEGORY_ORDER[groupId]
  );

  return orderedSubcategories
    .filter((id) => groups.has(id))
    .map((id) => ({
      id,
      activityTypes: groups.get(id)!,
    }));
}
