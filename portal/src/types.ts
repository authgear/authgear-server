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

// OAuthSSOConfig
export const oauthSSOProviderTypes = [
  "apple",
  "google",
  "facebook",
  "linkedin",
  "azureadv2",
] as const;
export type OAuthSSOProviderType = typeof oauthSSOProviderTypes[number];
export function isOAuthSSOProviderType(
  value?: string
): value is OAuthSSOProviderType {
  return value == null
    ? false
    : oauthSSOProviderTypes.includes(value as OAuthSSOProviderType);
}

export interface OAuthSSOProviderConfig {
  alias?: string;
  type: OAuthSSOProviderType;
  client_id?: string;
  tenant?: string;
  key_id?: string;
  team_id?: string;
}

interface OAuthSSOConfig {
  providers?: OAuthSSOProviderConfig[];
}

// IdentityConflictConfig
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
  oauth?: OAuthSSOConfig;
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

// secret config
export const secretConfigKeyList = [
  "db",
  "redis",
  "admin-api.auth",
  "sso.oauth.client",
  "mail.smtp",
  "sms.twilio",
  "sms.nexmo",
  "oidc",
  "csrf",
  "webhook",
] as const;
export type SecretConfigKey = typeof secretConfigKeyList[number];

// item with different key has different schema

interface DbSecretItem {
  key: "db";
  data: Record<string, unknown>;
}

interface RedisSecretItem {
  key: "redis";
  data: Record<string, unknown>;
}

interface AdminApiSecretItem {
  key: "admin-api.auth";
  data: Record<string, unknown>;
}

// sso.oauth.client

export interface OAuthClientCredentialItem {
  alias: string;
  client_secret: string;
}

interface OAuthClientCredentials {
  items: OAuthClientCredentialItem[];
}

export interface OAuthSecretItem {
  key: "sso.oauth.client";
  data: OAuthClientCredentials;
}

interface SmtpSecretItem {
  key: "mail.smtp";
  data: Record<string, unknown>;
}

interface TwilioSecretItem {
  key: "sms.twilio";
  data: Record<string, unknown>;
}

interface NexmoSecretItem {
  key: "sms.nexmo";
  data: Record<string, unknown>;
}

interface OidcSecretItem {
  key: "oidc";
  data: Record<string, unknown>;
}

interface CsrfSecretItem {
  key: "csrf";
  data: Record<string, unknown>;
}

interface WebhookSecretItem {
  key: "webhook";
  data: Record<string, unknown>;
}

// union type
type SecretItem =
  | DbSecretItem
  | RedisSecretItem
  | AdminApiSecretItem
  | OAuthSecretItem
  | SmtpSecretItem
  | TwilioSecretItem
  | NexmoSecretItem
  | OidcSecretItem
  | CsrfSecretItem
  | WebhookSecretItem;

export interface PortalAPISecretConfig {
  secrets: SecretItem[];
}

export interface PortalAPIApp {
  id: string;
  rawAppConfig?: PortalAPIAppConfig;
  effectiveAppConfig?: PortalAPIAppConfig;
  secretConfig?: PortalAPISecretConfig;
}
