declare type GQL_AuditLogData = unknown;

declare type GQL_DateTime = string;

declare interface GQL_IdentityClaims extends Record<string, unknown> {
  email?: string;
  preferred_username?: string;
  phone_number?: string;
}

declare type GQL_AuthenticatorClaims = Record<string, unknown>;

// If a .d.ts has a import statement, it becomes a normal module instead of a ambient module.
// So we have to use import() to import types here.
// eslint-disable-next-line no-undef
declare type GQL_AppConfig = import("./types").PortalAPIAppConfig;

// eslint-disable-next-line no-undef
declare type GQL_SecretConfig = import("./types").PortalAPISecretConfig;
