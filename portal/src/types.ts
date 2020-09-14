// LoginIDKeyConfig

export type LoginIDKeyType = "raw" | "email" | "phone" | "username";

export interface LoginIDKeyConfig {
  key: string;
  maximum?: number;
  type: LoginIDKeyType;
}

// LoginIDTypesConfig
interface LoginIDEmailConfig {
  block_plus_sign?: boolean;
  case_sensitive?: boolean;
  ignore_dot_sign?: boolean;
}

interface LoginIDUsernameConfig {
  block_reserved_usernames?: boolean;
  excluded_keywords?: string[];
  ascii_only?: boolean;
  case_sensitive?: boolean;
}

interface LoginIDTypesConfig {
  email?: LoginIDEmailConfig;
  username?: LoginIDUsernameConfig;
}

// LoginIDConfig
interface LoginIDConfig {
  keys?: LoginIDKeyConfig[];
  types?: LoginIDTypesConfig;
}

export const promotionConflictBehaviours = ["error", "login"] as const;
export type PromotionConflictBehaviour = typeof promotionConflictBehaviours[number];
export const isPromotionConflictBehaviour = (
  value: unknown
): value is PromotionConflictBehaviour => {
  if (typeof value !== "string") {
    return false;
  }
  return promotionConflictBehaviours.some(
    (behaviour: string) => behaviour === value
  );
};

interface IdentityConflictConfig {
  additionalProperties?: boolean;
  promotion?: PromotionConflictBehaviour;
}

interface IdentityConfig {
  login_id?: LoginIDConfig;
  on_conflict?: IdentityConflictConfig;
}

// AuthenticatorConfig

interface AuthenticatorConfig {
  oob_otp?: Record<string, unknown>;
  password?: Record<string, unknown>;
  totp?: Record<string, unknown>;
}

export const primaryAuthenticatorTypes = ["password", "oob_otp"] as const;
export type PrimaryAuthenticatorType = typeof primaryAuthenticatorTypes[number];
export function isPrimaryAuthenticatorType(
  type: any
): type is PrimaryAuthenticatorType {
  return primaryAuthenticatorTypes.includes(type);
}

export const secondaryAuthenticatorTypes = [
  "password",
  "oob_otp",
  "totp",
] as const;
export type SecondaryAuthenticatorType = typeof secondaryAuthenticatorTypes[number];
export function isSecondaryAuthenticatorType(
  type: any
): type is SecondaryAuthenticatorType {
  return secondaryAuthenticatorTypes.includes(type);
}

export const identityTypes = ["login_id", "oauth", "anonymous"] as const;
export type IdentityType = typeof identityTypes[number];

interface AuthenticationConfig {
  identities?: IdentityType[];
  primary_authenticators?: PrimaryAuthenticatorType[];
  secondary_authenticators?: SecondaryAuthenticatorType[];
}

export interface VerificationClaimConfig {
  enabled?: boolean;
  required?: boolean;
}

interface VerificationClaimsConfig {
  email?: VerificationClaimConfig;
  phone_number?: VerificationClaimConfig;
}

export const verificationCriteriaList = ["any", "all"] as const;
export type VerificationCriteria = typeof verificationCriteriaList[number];

// type alias of integer in JSON schema
type DurationSeconds = number;

interface VerificationConfig {
  claims?: VerificationClaimsConfig;
  criteria?: VerificationCriteria;
  code_expiry_seconds?: DurationSeconds;
}

export interface PortalAPIAppConfig {
  identity?: IdentityConfig;
  authenticator?: AuthenticatorConfig;
  authentication?: AuthenticationConfig;
  verification?: VerificationConfig;
}

export interface PortalAPIApp {
  id: string;
  rawAppConfig?: PortalAPIAppConfig;
  effectiveAppConfig?: PortalAPIAppConfig;
  secretConfig?: Record<string, unknown>;
}
