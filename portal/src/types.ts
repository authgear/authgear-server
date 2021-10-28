// HTTPConfig

export interface HTTPConfig {
  public_origin?: string;
  allowed_origins?: string[];
  cookie_prefix?: string;
  cookie_domain?: string;
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
  force_change?: boolean;
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

export interface DeviceTokenConfig {
  disabled?: boolean;
}

export interface AuthenticationConfig {
  identities?: IdentityType[];
  primary_authenticators?: PrimaryAuthenticatorType[];
  secondary_authenticators?: SecondaryAuthenticatorType[];
  secondary_authentication_mode?: SecondaryAuthenticationMode;
  recovery_code?: RecoveryCodeConfig;
  device_token?: DeviceTokenConfig;
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
export interface PhoneInputConfig {
  allowlist?: string[];
  pinned_list?: string[];
  preselect_by_ip_disabled?: boolean;
}

export interface UIConfig {
  phone_input?: PhoneInputConfig;
  dark_theme_disabled?: boolean;
  watermark_disabled?: boolean;
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

export interface StandardAttributesPopulationConfig {
  strategy?: "none" | "on_signup";
}

export type AccessControlLevelString = "hidden" | "readonly" | "readwrite";

export interface StandardAttributesAccessControl {
  end_user: AccessControlLevelString;
  bearer: AccessControlLevelString;
  portal_ui: AccessControlLevelString;
}

export interface StandardAttributesAccessControlConfig {
  pointer: string;
  access_control: StandardAttributesAccessControl;
}

export interface StandardAttributesConfig {
  population?: StandardAttributesPopulationConfig;
  access_control?: StandardAttributesAccessControlConfig[];
}

export interface UserProfileConfig {
  standard_attributes?: StandardAttributesConfig;
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
  user_profile?: UserProfileConfig;
}

export interface OAuthClientSecret {
  alias: string;
  clientSecret: string;
}

export interface WebhookSecret {
  secret?: string | null;
}

export interface AdminAPISecret {
  keyID: string;
  createdAt?: string | null;
  publicKeyPEM: string;
  privateKeyPEM?: string | null;
}

export interface SMTPSecret {
  host: string;
  port: number;
  username: string;
  password?: string | null;
}

export interface PortalAPISecretConfig {
  oauthClientSecrets?: OAuthClientSecret[] | null;
  webhookSecret?: WebhookSecret | null;
  adminAPISecrets?: AdminAPISecret[] | null;
  smtpSecret?: SMTPSecret | null;
}

export interface PortalAPIApp {
  id: string;
  rawAppConfig?: PortalAPIAppConfig;
  effectiveAppConfig?: PortalAPIAppConfig;
  secretConfig?: PortalAPISecretConfig;
}

export interface PortalAPIFeatureConfig {
  identity?: IdentityFeatureConfig;
  authentication?: AuthenticationFeatureConfig;
  custom_domain?: CustomDomainFeatureConfig;
  ui?: UIFeatureConfig;
  oauth?: OAuthFeatureConfig;
  hook?: HookFeatureConfig;
  audit_log?: AuditLogFeatureConfig;
}

export interface AuthenticationFeatureConfig {
  secondary_authenticators?: AuthenticatorsFeatureConfig;
}

export interface AuthenticatorsFeatureConfig {
  oob_otp_sms?: AuthenticatorOOBOTBSMSFeatureConfig;
}

export interface AuthenticatorOOBOTBSMSFeatureConfig {
  disabled?: boolean;
}

export interface IdentityFeatureConfig {
  login_id?: LoginIDFeatureConfig;
  oauth?: OAuthSSOFeatureConfig;
}

export interface LoginIDFeatureConfig {
  types?: LoginIDTypesFeatureConfig;
}

export interface LoginIDTypesFeatureConfig {
  phone?: LoginIDPhoneFeatureConfig;
}

export interface LoginIDPhoneFeatureConfig {
  disabled?: boolean;
}

export interface OAuthSSOFeatureConfig {
  maximum_providers?: number;
  providers?: OAuthSSOProvidersFeatureConfig;
}
export interface OAuthSSOProvidersFeatureConfig {
  google?: OAuthSSOProviderFeatureConfig;
  facebook?: OAuthSSOProviderFeatureConfig;
  linkedin?: OAuthSSOProviderFeatureConfig;
  azureadv2?: OAuthSSOProviderFeatureConfig;
  adfs?: OAuthSSOProviderFeatureConfig;
  apple?: OAuthSSOProviderFeatureConfig;
  wechat?: OAuthSSOProviderFeatureConfig;
}

export interface OAuthSSOProviderFeatureConfig {
  disabled?: boolean;
}

export interface CustomDomainFeatureConfig {
  disabled?: boolean;
}

export interface UIFeatureConfig {
  white_labeling?: WhiteLabelingFeatureConfig;
}

export interface WhiteLabelingFeatureConfig {
  disabled?: boolean;
}

export interface OAuthFeatureConfig {
  client?: OAuthClientFeatureConfig;
}

export interface OAuthClientFeatureConfig {
  maximum?: number;
}

export interface HookFeatureConfig {
  blocking_handler?: BlockingHookHandlerFeatureConfig;
  non_blocking_handler?: NonBlockingHookHandlerFeatureConfig;
}

export interface BlockingHookHandlerFeatureConfig {
  maximum?: number;
}

export interface NonBlockingHookHandlerFeatureConfig {
  maximum?: number;
}

export interface AuditLogFeatureConfig {
  retrieval_days?: number;
}

export interface StandardAttributes {
  email?: string;
  email_verified?: boolean;
  phone_number?: string;
  phone_number_verified?: boolean;
  preferred_username?: string;
  family_name?: string;
  given_name?: string;
  middle_name?: string;
  name?: string;
  nickname?: string;
  picture?: string;
  profile?: string;
  website?: string;
  gender?: string;
  birthdate?: string;
  zoneinfo?: string;
  locale?: string;
  address?: StandardAttributesAddress;
  updated_at?: number;
}

export interface StandardAttributesAddress {
  formatted?: string;
  street_address?: string;
  locality?: string;
  region?: string;
  postal_code?: string;
  country?: string;
}

export interface Identity {
  id: string;
  claims: IdentityClaims;
}

export interface IdentityClaims {
  email?: string;
  preferred_username?: string;
  phone_number?: string;
}

export type SessionType = "IDP" | "OFFLINE_GRANT";

export interface Session {
  id: string;
  type: SessionType;
  lastAccessedAt: string;
  lastAccessedByIP: string;
  displayName: string;
}
