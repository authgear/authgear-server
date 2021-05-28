// HTTPConfig

export interface HTTPConfig {
  public_origin?: string;
  allowed_origins?: string[];
  cookie_prefix?: string;
}

// LoginIDKeyConfig

export const loginIDKeyTypes = ["email", "phone", "username"] as const;
export type LoginIDKeyType = typeof loginIDKeyTypes[number];

export interface LoginIDKeyConfig {
  key: string;
  type: LoginIDKeyType;
  maximum?: number;
  modify_disabled?: boolean;
}

// LoginIDTypesConfig
export interface LoginIDEmailConfig {
  block_plus_sign?: boolean;
  case_sensitive?: boolean;
  ignore_dot_sign?: boolean;
  domain_blocklist_enabled?: boolean;
  domain_allowlist_enabled?: boolean;
  block_free_email_provider_domains?: boolean;
}

export interface LoginIDUsernameConfig {
  block_reserved_usernames?: boolean;
  exclude_keywords_enabled?: boolean;
  ascii_only?: boolean;
  case_sensitive?: boolean;
}

export interface LoginIDTypesConfig {
  email?: LoginIDEmailConfig;
  username?: LoginIDUsernameConfig;
}

// LoginIDConfig
export interface LoginIDConfig {
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
  "adfs",
  "wechat",
] as const;
export type OAuthSSOProviderType = typeof oauthSSOProviderTypes[number];
export const oauthSSOWeChatAppType = ["mobile", "web"] as const;
export type OAuthSSOWeChatAppType = typeof oauthSSOWeChatAppType[number];
export interface OAuthSSOProviderConfig {
  alias: string;
  type: OAuthSSOProviderType;
  modify_disabled?: boolean;
  client_id?: string;
  tenant?: string;
  key_id?: string;
  team_id?: string;
  app_type?: OAuthSSOWeChatAppType;
  account_id?: string;
  is_sandbox_account?: boolean;
  wechat_redirect_uris?: string[];
  discovery_document_endpoint?: string;
}
export const oauthSSOProviderItemKeys = [
  "apple",
  "google",
  "facebook",
  "linkedin",
  "azureadv2",
  "adfs",
  "wechat.mobile",
  "wechat.web",
] as const;
export type OAuthSSOProviderItemKey = typeof oauthSSOProviderItemKeys[number];

export const createOAuthSSOProviderItemKey = (
  type: OAuthSSOProviderType,
  appType?: OAuthSSOWeChatAppType
): OAuthSSOProviderItemKey => {
  return (
    !appType ? type : [type, appType].join(".")
  ) as OAuthSSOProviderItemKey;
};

export const parseOAuthSSOProviderItemKey = (
  itemKey: OAuthSSOProviderItemKey
): [OAuthSSOProviderType, OAuthSSOWeChatAppType?] => {
  const [type, appType] = itemKey.split(".");
  return [type as OAuthSSOProviderType, appType as OAuthSSOWeChatAppType];
};

export const isOAuthSSOProvider = (
  config: OAuthSSOProviderConfig,
  type: OAuthSSOProviderType,
  appType?: OAuthSSOWeChatAppType
): boolean => {
  return config.type === type && (!appType || config.app_type === appType);
};

export interface OAuthSSOConfig {
  providers?: OAuthSSOProviderConfig[];
}

// IdentityConflictConfig
export const promotionConflictBehaviours = ["error", "login"] as const;
export type PromotionConflictBehaviour =
  typeof promotionConflictBehaviours[number];
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

export interface IdentityConflictConfig {
  additionalProperties?: boolean;
  promotion?: PromotionConflictBehaviour;
}

export interface BiometricConfig {
  list_enabled?: boolean;
}

export interface IdentityConfig {
  login_id?: LoginIDConfig;
  oauth?: OAuthSSOConfig;
  biometric?: BiometricConfig;
  on_conflict?: IdentityConflictConfig;
}

// AuthenticatorConfig

export const passwordPolicyGuessableLevels = [0, 1, 2, 3, 4, 5] as const;
export type PasswordPolicyGuessableLevel =
  typeof passwordPolicyGuessableLevels[number];
export const isPasswordPolicyGuessableLevel = (
  value: unknown
): value is PasswordPolicyGuessableLevel => {
  if (typeof value !== "number") {
    return false;
  }
  return passwordPolicyGuessableLevels.some((level) => level === value);
};

export interface PasswordPolicyConfig {
  min_length?: number;
  uppercase_required?: boolean;
  lowercase_required?: boolean;
  digit_required?: boolean;
  symbol_required?: boolean;
  minimum_guessable_level?: PasswordPolicyGuessableLevel;
  excluded_keywords?: string[];
  history_size?: number;
  history_days?: number;
}

export interface AuthenticatorPasswordConfig {
  policy?: PasswordPolicyConfig;
}

export interface AuthenticatorConfig {
  oob_otp?: Record<string, unknown>;
  password?: AuthenticatorPasswordConfig;
  totp?: Record<string, unknown>;
}

export const primaryAuthenticatorTypes = [
  "password",
  "oob_otp_email",
  "oob_otp_sms",
] as const;
export type PrimaryAuthenticatorType = typeof primaryAuthenticatorTypes[number];
export function isPrimaryAuthenticatorType(
  type: any
): type is PrimaryAuthenticatorType {
  return primaryAuthenticatorTypes.includes(type);
}

export const secondaryAuthenticatorTypes = [
  "password",
  "oob_otp_email",
  "oob_otp_sms",
  "totp",
] as const;
export type SecondaryAuthenticatorType =
  typeof secondaryAuthenticatorTypes[number];
export function isSecondaryAuthenticatorType(
  type: any
): type is SecondaryAuthenticatorType {
  return secondaryAuthenticatorTypes.includes(type);
}

export const identityTypes = [
  "login_id",
  "oauth",
  "anonymous",
  "biometric",
] as const;
export type IdentityType = typeof identityTypes[number];

export const secondaryAuthenticationModes = [
  "if_requested",
  "if_exists",
  "required",
] as const;
export type SecondaryAuthenticationMode =
  typeof secondaryAuthenticationModes[number];

export interface RecoveryCodeConfig {
  count?: number;
  list_enabled?: boolean;
}

export interface AuthenticationConfig {
  identities?: IdentityType[];
  primary_authenticators?: PrimaryAuthenticatorType[];
  secondary_authenticators?: SecondaryAuthenticatorType[];
  secondary_authentication_mode?: SecondaryAuthenticationMode;
  recovery_code?: RecoveryCodeConfig;
}

export interface VerificationClaimConfig {
  enabled?: boolean;
  required?: boolean;
}

export interface VerificationClaimsConfig {
  email?: VerificationClaimConfig;
  phone_number?: VerificationClaimConfig;
}

export const verificationCriteriaList = ["any", "all"] as const;
export type VerificationCriteria = typeof verificationCriteriaList[number];

// type alias of integer in JSON schema
export type DurationSeconds = number;

export interface VerificationConfig {
  claims?: VerificationClaimsConfig;
  criteria?: VerificationCriteria;
  code_expiry_seconds?: DurationSeconds;
}

// UIConfig
export interface UICountryCallingCodeConfig {
  allowlist?: string[];
  pinned_list?: string[];
}

export interface UIConfig {
  country_calling_code?: UICountryCallingCodeConfig;
  dark_theme_disabled?: boolean;
  default_client_uri?: string;
  default_redirect_uri?: string;
  default_post_logout_redirect_uri?: string;
}

// LocalizationConfig
export interface LocalizationConfig {
  supported_languages?: string[];
  fallback_language?: string;
}

// ForgotPasswordConfig
export interface ForgotPasswordConfig {
  enabled?: boolean;
  reset_code_expiry_seconds?: DurationSeconds;
}

// OAuthConfig
export interface OAuthClientConfig {
  name?: string;
  client_id: string;
  client_uri?: string;
  redirect_uris: string[];
  grant_types?: string[];
  response_types?: string[];
  post_logout_redirect_uris?: string[];
  access_token_lifetime_seconds?: number;
  refresh_token_lifetime_seconds?: number;
  refresh_token_idle_timeout_seconds?: number;
  refresh_token_idle_timeout_enabled?: boolean;
  issue_jwt_access_token?: boolean;
}

export interface OAuthConfig {
  clients?: OAuthClientConfig[];
}

// SessionConfig
export interface SessionConfig {
  cookie_domain?: string;
  cookie_non_persistent?: boolean;
  idle_timeout_enabled?: boolean;
  idle_timeout_seconds?: DurationSeconds;
  lifetime_seconds?: DurationSeconds;
}

export interface HookConfig {
  sync_hook_timeout_seconds?: number;
  sync_hook_total_timeout_seconds?: number;
  blocking_handlers?: BlockingHookHandlerConfig[];
  non_blocking_handlers?: NonBlockingHookHandlerConfig[];
}

export interface BlockingHookHandlerConfig {
  event: string;
  url: string;
}

export interface NonBlockingHookHandlerConfig {
  events: string[];
  url: string;
}

// PortalAPIAppConfig
export interface PortalAPIAppConfig {
  id: string;
  http?: HTTPConfig;
  identity?: IdentityConfig;
  authenticator?: AuthenticatorConfig;
  authentication?: AuthenticationConfig;
  verification?: VerificationConfig;
  ui?: UIConfig;
  localization?: LocalizationConfig;
  forgot_password?: ForgotPasswordConfig;
  oauth?: OAuthConfig;
  session?: SessionConfig;
  hook?: HookConfig;
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

export interface DbSecretItem {
  key: "db";
  data: Record<string, unknown>;
}

export interface RedisSecretItem {
  key: "redis";
  data: Record<string, unknown>;
}

export interface AdminApiSecretItem {
  key: "admin-api.auth";
  data: Record<string, unknown>;
}

// sso.oauth.client

export interface OAuthClientCredentialItem {
  alias: string;
  client_secret: string;
}

export interface OAuthClientCredentials {
  items: OAuthClientCredentialItem[];
}

export interface OAuthSecretItem {
  key: "sso.oauth.client";
  data: OAuthClientCredentials;
}

export interface SmtpSecretItem {
  key: "mail.smtp";
  data: Record<string, unknown>;
}

export interface TwilioSecretItem {
  key: "sms.twilio";
  data: Record<string, unknown>;
}

export interface NexmoSecretItem {
  key: "sms.nexmo";
  data: Record<string, unknown>;
}

export interface OidcSecretItem {
  key: "oidc";
  data: Record<string, unknown>;
}

export interface CsrfSecretItem {
  key: "csrf";
  data: Record<string, unknown>;
}

export interface WebhookSecretItem {
  key: "webhook";
  data: Record<string, unknown>;
}

// union type
export type SecretItem =
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
