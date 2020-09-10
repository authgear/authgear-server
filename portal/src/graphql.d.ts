declare type GQL_DateTime = string;

declare interface GQL_IdentityClaims extends Record<string, unknown> {
  email?: string;
  preferred_username?: string;
  phone_number?: string;
}

declare type GQL_AuthenticatorClaims = Record<string, unknown>;

declare type GQL_AppConfig = PortalAPIAppConfig;

declare type GQL_SecretConfig = Record<string, unknown>;
