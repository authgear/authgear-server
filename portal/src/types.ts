import { OAuthSsoProviderClientSecretInput } from "./graphql/portal/globalTypes.generated";

// type aliases in JSON schema
export type DurationString = string;
export type DurationSeconds = number;

// HTTPConfig

export interface HTTPConfig {
  public_origin?: string;
  allowed_origins?: string[];
  cookie_prefix?: string;
  cookie_domain?: string;
}

// RateLimitConfig

export interface RateLimitConfig {
  enabled?: boolean;
  period?: DurationString;
  burst?: number;
}

// LoginIDKeyConfig

export const loginIDKeyTypes = ["email", "phone", "username"] as const;
export type LoginIDKeyType = (typeof loginIDKeyTypes)[number];

export interface LoginIDKeyConfig {
  type: LoginIDKeyType;
  maximum?: number;
  create_disabled?: boolean;
  update_disabled?: boolean;
  delete_disabled?: boolean;
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
  "github",
  "linkedin",
  "azureadv2",
  "azureadb2c",
  "adfs",
  "wechat",
] as const;
export type OAuthSSOProviderType = (typeof oauthSSOProviderTypes)[number];
export const oauthSSOWeChatAppType = ["mobile", "web"] as const;
export type OAuthSSOWeChatAppType = (typeof oauthSSOWeChatAppType)[number];
export interface OAuthClaimConfig {
  required?: boolean;
}
export interface OAuthClaimsConfig {
  email?: OAuthClaimConfig;
}
export interface OAuthSSOProviderConfig {
  alias: string;
  type: OAuthSSOProviderType;
  create_disabled?: boolean;
  delete_disabled?: boolean;
  client_id?: string;
  tenant?: string;
  key_id?: string;
  team_id?: string;
  app_type?: OAuthSSOWeChatAppType;
  account_id?: string;
  is_sandbox_account?: boolean;
  wechat_redirect_uris?: string[];
  discovery_document_endpoint?: string;
  policy?: string;
  claims?: OAuthClaimsConfig;
}
export const oauthSSOProviderItemKeys = [
  "apple",
  "google",
  "facebook",
  "github",
  "linkedin",
  "azureadv2",
  "azureadb2c",
  "adfs",
  "wechat.mobile",
  "wechat.web",
] as const;
export type OAuthSSOProviderItemKey = (typeof oauthSSOProviderItemKeys)[number];

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
  (typeof promotionConflictBehaviours)[number];
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
  (typeof passwordPolicyGuessableLevels)[number];
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
  alphabet_required?: boolean;
  digit_required?: boolean;
  symbol_required?: boolean;
  minimum_guessable_level?: PasswordPolicyGuessableLevel;
  excluded_keywords?: string[];
  history_size?: number;
  history_days?: number;
}

export interface PasswordExpiryConfig {
  force_change?: PasswordExpiryForceChangeConfig;
}

export interface PasswordExpiryForceChangeConfig {
  enabled?: boolean;
  duration_since_last_update?: DurationString;
}

export interface AuthenticatorOOBConfig {
  email?: AuthenticatorOOBEmailConfig;
  sms?: AuthenticatorOOBSMSConfig;
}

export interface AuthenticatorOOBEmailConfig {
  maximum?: number;
  email_otp_mode?: AuthenticatorEmailOTPMode;
  valid_periods?: AuthenticatorValidPeriods;
}

export interface AuthenticatorOOBSMSConfig {
  maximum?: number;
  phone_otp_mode?: AuthenticatorPhoneOTPMode;
  valid_periods?: AuthenticatorValidPeriods;
}
export interface AuthenticatorValidPeriods {
  link?: DurationString;
  code?: DurationString;
}

export interface AuthenticatorPasswordConfig {
  force_change?: boolean;
  policy?: PasswordPolicyConfig;
  expiry?: PasswordExpiryConfig;
}

export interface AuthenticatorConfig {
  oob_otp?: AuthenticatorOOBConfig;
  password?: AuthenticatorPasswordConfig;
  totp?: Record<string, unknown>;
}

export type AuthenticationLockoutType = "per_user" | "per_user_per_ip";

export interface AuthenticationLockoutMethodConfig {
  enabled: boolean;
}

export interface AuthenticationLockoutConfig {
  max_attempts?: number;
  history_duration?: DurationString;
  minimum_duration?: DurationString;
  maximum_duration?: DurationString;
  backoff_factor?: number;
  lockout_type?: AuthenticationLockoutType;
  password?: AuthenticationLockoutMethodConfig;
  totp?: AuthenticationLockoutMethodConfig;
  oob_otp?: AuthenticationLockoutMethodConfig;
  recovery_code?: AuthenticationLockoutMethodConfig;
}

export const primaryAuthenticatorTypes = [
  "password",
  "oob_otp_email",
  "oob_otp_sms",
  "passkey",
] as const;
export type PrimaryAuthenticatorType =
  (typeof primaryAuthenticatorTypes)[number];
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
  (typeof secondaryAuthenticatorTypes)[number];
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
  "passkey",
  "siwe",
] as const;
export type IdentityType = (typeof identityTypes)[number];

export const secondaryAuthenticationModes = [
  "disabled",
  "if_exists",
  "required",
] as const;
export type SecondaryAuthenticationMode =
  (typeof secondaryAuthenticationModes)[number];

export interface RecoveryCodeConfig {
  disabled?: boolean;
  count?: number;
  list_enabled?: boolean;
}

export interface MFAGlobalGracePeriodConfig {
  enabled?: boolean;
  endAt?: string;
}

export interface DeviceTokenConfig {
  disabled?: boolean;
}

export interface AuthenticationConfig {
  identities?: IdentityType[];
  primary_authenticators?: PrimaryAuthenticatorType[];
  secondary_authenticators?: SecondaryAuthenticatorType[];
  secondary_authentication_mode?: SecondaryAuthenticationMode;
  secondary_authentication_grace_period?: MFAGlobalGracePeriodConfig;
  recovery_code?: RecoveryCodeConfig;
  device_token?: DeviceTokenConfig;
  rate_limits?: AuthenticationRateLimitsConfig;
  lockout?: AuthenticationLockoutConfig;
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
export type VerificationCriteria = (typeof verificationCriteriaList)[number];

export const authenticatorEmailOTPModeList = ["code", "login_link"] as const;
export type AuthenticatorEmailOTPMode =
  (typeof authenticatorEmailOTPModeList)[number];

export const authenticatorPhoneOTPModeList = [
  "sms",
  "whatsapp_sms",
  "whatsapp",
] as const;
export type AuthenticatorPhoneOTPMode =
  (typeof authenticatorPhoneOTPModeList)[number];

export const authenticatorPhoneOTPSMSModeList = ["sms", "whatsapp_sms"];

export interface AuthenticationRateLimitsConfig {
  oob_otp?: AuthenticationRateLimitsOOBOTPConfig;
}

export interface AuthenticationRateLimitsOOBOTPConfig {
  email?: AuthenticationRateLimitsEmailConfig;
  sms?: AuthenticationRateLimitsSMSConfig;
}

export interface AuthenticationRateLimitsEmailConfig {
  trigger_cooldown?: DurationString;
  max_failed_attempts_revoke_otp?: number;
}

export interface AuthenticationRateLimitsSMSConfig {
  trigger_cooldown?: DurationString;
  max_failed_attempts_revoke_otp?: number;
}

export interface VerificationConfig {
  claims?: VerificationClaimsConfig;
  criteria?: VerificationCriteria;
  code_valid_period?: DurationString;
  rate_limits?: VerificationRateLimitsConfig;
}

export interface VerificationRateLimitsConfig {
  email?: VerificationRateLimitsEmailConfig;
  sms?: VerificationRateLimitsSMSConfig;
}

export interface VerificationRateLimitsEmailConfig {
  trigger_cooldown?: DurationString;
  max_failed_attempts_revoke_otp?: number;
  trigger_per_user?: RateLimitConfig;
}

export interface VerificationRateLimitsSMSConfig {
  trigger_cooldown?: DurationString;
  max_failed_attempts_revoke_otp?: number;
  trigger_per_user?: RateLimitConfig;
}

export interface MessagingConfig {
  rate_limits?: MessagingRateLimitsConfig;
  template_customization_disabled?: boolean;
}

export interface MessagingRateLimitsConfig {
  sms?: RateLimitConfig;
  sms_per_ip?: RateLimitConfig;
  sms_per_target?: RateLimitConfig;
  email?: RateLimitConfig;
  email_per_ip?: RateLimitConfig;
  email_per_target?: RateLimitConfig;
}

// UIConfig
export interface PhoneInputConfig {
  allowlist?: string[];
  pinned_list?: string[];
  preselect_by_ip_disabled?: boolean;
}

export interface UIConfig {
  implementation?: UIImplementation;
  signup_login_flow_enabled?: boolean;
  phone_input?: PhoneInputConfig;
  dark_theme_disabled?: boolean;
  light_theme_disabled?: boolean;
  watermark_disabled?: boolean;
  default_client_uri?: string;
  default_brand_page_redirect_uri?: string;
  default_redirect_uri?: string;
  default_post_logout_redirect_uri?: string;
  forgot_password?: UIForgotPasswordConfig;
}

export interface UIForgotPasswordConfig {
  phone?: AccountRecoveryChannel[];
  email?: AccountRecoveryChannel[];
}

export interface AccountRecoveryChannel {
  channel: AccountRecoveryCodeChannel;
  otp_form: AccountRecoveryCodeForm;
}

export enum AccountRecoveryCodeChannel {
  Email = "email",
  SMS = "sms",
  Whatsapp = "whatsapp",
}

export enum AccountRecoveryCodeForm {
  Link = "link",
  Code = "code",
}

// LocalizationConfig
export interface LocalizationConfig {
  supported_languages?: string[];
  fallback_language?: string;
}

// ForgotPasswordConfig
export interface ForgotPasswordConfig {
  enabled?: boolean;
  valid_periods?: ForgotPasswordValidPeriods;
}

export interface ForgotPasswordValidPeriods {
  link?: DurationString;
  code?: DurationString;
}

export const applicationTypes = [
  "spa",
  "traditional_webapp",
  "native",
  "confidential",
  "third_party_app",
] as const;
export type ApplicationType = (typeof applicationTypes)[number];

// OAuthConfig
export interface OAuthClientConfig {
  name?: string;
  client_id: string;
  client_uri?: string;
  client_name?: string;
  x_application_type?: ApplicationType;
  x_max_concurrent_session?: number;
  redirect_uris: string[];
  grant_types?: string[];
  response_types?: string[];
  post_logout_redirect_uris?: string[];
  access_token_lifetime_seconds?: number;
  refresh_token_lifetime_seconds?: number;
  refresh_token_idle_timeout_seconds?: number;
  refresh_token_idle_timeout_enabled?: boolean;
  issue_jwt_access_token?: boolean;
  policy_uri?: string;
  tos_uri?: string;
  x_custom_ui_uri?: string;
  x_app2app_enabled?: boolean;
  x_app2app_insecure_device_key_binding_enabled?: boolean;
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

export interface UserProfileAttributesAccessControl {
  end_user: AccessControlLevelString;
  bearer: AccessControlLevelString;
  portal_ui: AccessControlLevelString;
}

export interface StandardAttributesAccessControlConfig {
  pointer: string;
  access_control: UserProfileAttributesAccessControl;
}

export interface StandardAttributesConfig {
  population?: StandardAttributesPopulationConfig;
  access_control?: StandardAttributesAccessControlConfig[];
}

export interface CustomAttributesConfig {
  attributes?: CustomAttributesAttributeConfig[];
}

export interface CustomAttributesAttributeConfig {
  id: string;
  pointer: string;
  type: CustomAttributeType;
  access_control: UserProfileAttributesAccessControl;
  minimum?: number;
  maximum?: number;
  enum?: string[];
}

export const customAttributeTypes = [
  "string",
  "number",
  "integer",
  "enum",
  "phone_number",
  "email",
  "url",
  "country_code",
] as const;
export type CustomAttributeType = (typeof customAttributeTypes)[number];
export function isCustomAttributeType(v: unknown): v is CustomAttributeType {
  // @ts-expect-error
  return typeof v === "string" && customAttributeTypes.includes(v);
}

export interface UserProfileConfig {
  standard_attributes?: StandardAttributesConfig;
  custom_attributes?: CustomAttributesConfig;
}

// Web3 Configs
export interface SIWEConfig {
  networks?: string[];
}

export interface NFTConfig {
  collections?: string[];
}

export interface Web3Config {
  nft?: NFTConfig;
  siwe?: SIWEConfig;
}

// PortalAPIAppConfig
export interface PortalAPIAppConfig {
  id: string;
  http?: HTTPConfig;
  identity?: IdentityConfig;
  authenticator?: AuthenticatorConfig;
  authentication?: AuthenticationConfig;
  verification?: VerificationConfig;
  messaging?: MessagingConfig;
  ui?: UIConfig;
  localization?: LocalizationConfig;
  forgot_password?: ForgotPasswordConfig;
  oauth?: OAuthConfig;
  session?: SessionConfig;
  hook?: HookConfig;
  web3?: Web3Config;
  user_profile?: UserProfileConfig;
  account_deletion?: AccountDeletionConfig;
  account_anonymization?: AccountAnonymizationConfig;
  google_tag_manager?: GoogleTagManagerConfig;
  bot_protection?: BotProtectionConfig;
}

export interface OAuthSSOProviderClientSecret {
  alias: string;
  clientSecret?: string | null;
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

export interface OAuthClientSecretKey {
  keyID: string;
  createdAt?: string | null;
  key: string;
}

export interface OAuthClientSecret {
  clientID: string;
  keys?: OAuthClientSecretKey[] | null;
}

export interface BotProtectionProviderSecret {
  secretKey?: string | null;
  type: string;
}

export interface PortalAPISecretConfig {
  oauthSSOProviderClientSecrets?: OAuthSSOProviderClientSecret[] | null;
  webhookSecret?: WebhookSecret | null;
  adminAPISecrets?: AdminAPISecret[] | null;
  smtpSecret?: SMTPSecret | null;
  oauthClientSecrets?: OAuthClientSecret[] | null;
  botProtectionProviderSecret?: BotProtectionProviderSecret | null;
}

export interface OAuthSSOProviderClientSecretUpdateInstructionDataItem
  extends OAuthSsoProviderClientSecretInput {}

export interface SSOProviderFormSecretViewModel {
  originalAlias: string | null;
  newAlias: string;
  newClientSecret: string | null;
}

export interface OAuthSSOProviderClientSecretUpdateInstruction {
  action: string;
  data?: OAuthSSOProviderClientSecretUpdateInstructionDataItem[] | null;
}

export interface SMTPSecretUpdateInstructionData {
  host: string;
  port: number;
  username: string;
  password: string | null;
}

export interface SMTPSecretUpdateInstruction {
  action: string;
  data?: SMTPSecretUpdateInstructionData | null;
}

export interface BotProtectionProviderSecretUpdateInstructionData {
  secretKey: string | null;
  type: BotProtectionProviderType;
}
export interface BotProtectionProviderSecretUpdateInstruction {
  action: string;
  data?: BotProtectionProviderSecretUpdateInstructionData | null;
}

export interface OAuthClientSecretsUpdateInstructionGenerateData {
  clientID: string;
}

export interface OAuthClientSecretsUpdateInstructionCleanupData {
  keepClientIDs: string[];
}

export interface OAuthClientSecretsUpdateInstruction {
  action: string;
  generateData?: OAuthClientSecretsUpdateInstructionGenerateData | null;
  cleanupData?: OAuthClientSecretsUpdateInstructionCleanupData | null;
}

export interface AdminAPIAuthKeyDeleteDataInput {
  keyID: string;
}

export interface AdminApiAuthKeyUpdateInstruction {
  action: string;
  deleteData?: AdminAPIAuthKeyDeleteDataInput | null;
}

export interface PortalAPISecretConfigUpdateInstruction {
  oauthSSOProviderClientSecrets?: OAuthSSOProviderClientSecretUpdateInstruction | null;
  smtpSecret?: SMTPSecretUpdateInstruction | null;
  oauthClientSecrets?: OAuthClientSecretsUpdateInstruction | null;
  adminAPIAuthKey?: AdminApiAuthKeyUpdateInstruction | null;
  botProtectionProviderSecret?: BotProtectionProviderSecretUpdateInstruction | null;
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
  authenticator?: AuthenticatorFeatureConfig;
  custom_domain?: CustomDomainFeatureConfig;
  ui?: UIFeatureConfig;
  oauth?: OAuthFeatureConfig;
  hook?: HookFeatureConfig;
  audit_log?: AuditLogFeatureConfig;
  google_tag_manager?: GoogleTagManagerFeatureConfig;
  collaborator?: CollaboratorFeatureConfig;
  web3?: Web3FeatureConfig;
}

export interface AuthenticatorFeatureConfig {
  password?: AuthenticatorPasswordFeatureConfig;
}

export interface AuthenticatorPasswordFeatureConfig {
  policy?: PasswordPolicyFeatureConfig;
}

export interface PasswordPolicyFeatureConfig {
  minimum_guessable_level?: PasswordPolicyItemFeatureConfig;
  excluded_keywords?: PasswordPolicyItemFeatureConfig;
  history?: PasswordPolicyItemFeatureConfig;
}

export interface PasswordPolicyItemFeatureConfig {
  disabled?: boolean;
}

export interface CollaboratorFeatureConfig {
  maximum?: number;
}

export interface GoogleTagManagerFeatureConfig {
  disabled?: boolean;
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

export interface Web3NFTFeatureConfig {
  maximum?: number;
}

export interface Web3FeatureConfig {
  nft?: Web3NFTFeatureConfig;
}

export interface IdentityFeatureConfig {
  login_id?: LoginIDFeatureConfig;
  oauth?: OAuthSSOFeatureConfig;
  biometric?: BiometricFeatureConfig;
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
  github?: OAuthSSOProviderFeatureConfig;
  linkedin?: OAuthSSOProviderFeatureConfig;
  azureadv2?: OAuthSSOProviderFeatureConfig;
  azureadb2c?: OAuthSSOProviderFeatureConfig;
  adfs?: OAuthSSOProviderFeatureConfig;
  apple?: OAuthSSOProviderFeatureConfig;
  wechat?: OAuthSSOProviderFeatureConfig;
}

export interface OAuthSSOProviderFeatureConfig {
  disabled?: boolean;
}

export interface BiometricFeatureConfig {
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
  custom_ui_enabled: boolean;
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

export interface AccountDeletionConfig {
  scheduled_by_end_user_enabled?: boolean;
  grace_period_days?: number;
}

export interface AccountAnonymizationConfig {
  grace_period_days?: number;
}

export interface GoogleTagManagerConfig {
  container_id?: string;
}

export interface BotProtectionProviderConfig {
  site_key?: string; // assume all provides have site_key for now
  type?: BotProtectionProviderType;
}

export type BotProtectionRiskMode = "never" | "always";
export interface BotProtectionRequirementsObject {
  mode?: BotProtectionRiskMode;
}
export interface BotProtectionRequirements {
  signup_or_login?: BotProtectionRequirementsObject;
  account_recovery?: BotProtectionRequirementsObject;
  password?: BotProtectionRequirementsObject;
  oob_otp_email?: BotProtectionRequirementsObject;
  oob_otp_sms?: BotProtectionRequirementsObject;
}

export interface BotProtectionConfig {
  enabled?: boolean;
  provider?: BotProtectionProviderConfig;
  requirements?: BotProtectionRequirements;
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

export type CustomAttributes = Record<string, unknown>;

export interface StandardAttributesAddress {
  formatted?: string;
  street_address?: string;
  locality?: string;
  region?: string;
  postal_code?: string;
  country?: string;
}

export interface NFTContract {
  name: string;
  address: string;
  type: string;
}

export interface TransactionIdentifier {
  hash: string;
}

export interface BlockIdentifier {
  index: number;
  timestamp: Date;
}

export interface NFTToken {
  token_id: string;
  transaction_identifier: TransactionIdentifier;
  block_identifier: BlockIdentifier;
  balance: string;
}

export interface NFT {
  contract: NFTContract;
  tokens: NFTToken[];
}

export interface AccountIdentifier {
  address?: string;
}

export interface NetworkIdentifier {
  blockchain?: string;
  network?: string;
}
export interface Web3Account {
  account_identifier?: AccountIdentifier;
  network_identifier?: NetworkIdentifier;
  nfts?: NFT[];
}

export interface Web3Claims {
  accounts?: Web3Account[];
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
  userAgent?: string | null;
  clientID?: string | null;
}

export interface Authorization {
  id: string;
  clientID: string;
  createdAt: string;
  scopes: string[];
}

export interface TutorialStatusData {
  skipped?: boolean;
  progress: {
    authui?: boolean;
    customize_ui?: boolean;
    create_application?: boolean;
    sso?: boolean;
    invite?: boolean;
  };
}

export const botProtectionProviderTypes = [
  "cloudflare",
  "recaptchav2",
] as const;
export type BotProtectionProviderType =
  (typeof botProtectionProviderTypes)[number];

export type UIImplementation = "interaction" | "authflow" | "authflowv2";
