import MESSAGES from "./locale-data/en.json";
import { DEFAULT_TEMPLATE_LOCALE } from "./resources";

export interface SystemConfig {
  authgearAppID: string;
  authgearClientID: string;
  authgearEndpoint: string;
  authgearWebSDKSessionType: "cookie" | "refresh_token";
  isAuthgearOnce: boolean;
  authgearOnceLicenseKey: string;
  authgearOnceLicenseExpireAt: string;
  authgearOnceLicenseeEmail: string;
  sentryDSN: string;
  appHostSuffix: string;
  availableLanguages: string[];
  builtinLanguages: string[];
  translations: SystemConfigTranslations;
  searchEnabled: boolean;
  auditLogEnabled: boolean;
  gitCommitHash: string;
  analyticEnabled: boolean;
  analyticEpoch: string;
  gtmContainerID: string;
  uiImplementation: string;
  uiSettingsImplemenation: string;
}

export interface SystemConfigTranslations {
  en: Record<string, string>;
}

export interface PartialSystemConfig
  extends Partial<Omit<SystemConfig, "translations">> {
  translations?: Partial<SystemConfigTranslations>;
}

export const defaultSystemConfig: PartialSystemConfig = {
  translations: {
    en: MESSAGES,
  },
};

export function mergeSystemConfig(
  baseConfig: PartialSystemConfig,
  overlayConfig: PartialSystemConfig
): PartialSystemConfig {
  return {
    ...baseConfig,
    ...overlayConfig,
    translations: {
      en: {
        ...baseConfig.translations?.en,
        ...overlayConfig.translations?.en,
      },
    },
  };
}

export function instantiateSystemConfig(
  config: PartialSystemConfig
): SystemConfig {
  return {
    authgearAppID: config.authgearAppID ?? "",
    authgearClientID: config.authgearClientID ?? "",
    authgearEndpoint: config.authgearEndpoint ?? "",
    authgearWebSDKSessionType: config.authgearWebSDKSessionType ?? "cookie",
    isAuthgearOnce: config.isAuthgearOnce ?? false,
    authgearOnceLicenseKey: config.authgearOnceLicenseKey ?? "",
    authgearOnceLicenseExpireAt: config.authgearOnceLicenseExpireAt ?? "",
    authgearOnceLicenseeEmail: config.authgearOnceLicenseeEmail ?? "",
    sentryDSN: config.sentryDSN ?? "",
    appHostSuffix: config.appHostSuffix ?? "",
    availableLanguages: config.availableLanguages ?? [DEFAULT_TEMPLATE_LOCALE],
    builtinLanguages: config.builtinLanguages ?? [DEFAULT_TEMPLATE_LOCALE],
    translations: {
      en: config.translations?.en ?? {},
    },
    searchEnabled: config.searchEnabled ?? false,
    auditLogEnabled: config.auditLogEnabled ?? false,
    gitCommitHash: config.gitCommitHash ?? "",
    analyticEnabled: config.analyticEnabled ?? false,
    analyticEpoch: config.analyticEpoch ?? "",
    gtmContainerID: config.gtmContainerID ?? "",
    uiImplementation: config.uiImplementation ?? "authflowv2",
    uiSettingsImplemenation: config.uiSettingsImplemenation ?? "v2",
  };
}
