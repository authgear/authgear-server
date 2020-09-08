// LoginIDKeyConfig

export type LoginIDKeyType = "raw" | "email" | "phone" | "username";

interface VerificationLoginIDKeyConfig {
  enabled: boolean;
  required?: boolean;
}

export interface LoginIDKeyConfig {
  key?: string;
  maximum?: number;
  type: LoginIDKeyType;
  verification?: VerificationLoginIDKeyConfig;
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
  keys: LoginIDKeyConfig[];
  types: LoginIDTypesConfig;
}

interface IdentityConfig {
  login_id?: LoginIDConfig;
}

export interface PortalAPIAppConfig {
  identity?: IdentityConfig;
}

export interface PortalAPIApp {
  id: string;
  rawAppConfig?: PortalAPIAppConfig;
  effectiveAppConfig?: PortalAPIAppConfig;
  secretConfig?: Record<string, unknown>;
}
