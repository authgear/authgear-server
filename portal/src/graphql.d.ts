declare type GQL_AuditLogData = unknown;

declare type GQL_Date = string;

declare type GQL_DateTime = string;

declare interface GQL_IdentityClaims extends Record<string, unknown> {
  email?: string;
  preferred_username?: string;
  phone_number?: string;
}

declare type GQL_AuthenticatorClaims = Record<string, unknown>;

// If a .d.ts has a import statement, it becomes a normal module instead of a ambient module.
// So we have to use import() to import types here.

declare type GQL_AppConfig = import("./types").PortalAPIAppConfig;

declare type GQL_SecretConfig = import("./types").PortalAPISecretConfig;

declare type GQL_FeatureConfig = import("./types").PortalAPIFeatureConfig;

declare type GQL_UserStandardAttributes = import("./types").StandardAttributes;

declare type GQL_UserCustomAttributes = import("./types").CustomAttributes;

declare type GQL_TutorialStatusData = import("./types").TutorialStatusData;

declare type GQL_Web3Claims = import("./types").Web3Claims;

declare type GQL_StripeError = stripe.Error;
