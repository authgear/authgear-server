import { IPartialTheme, ITheme, createTheme } from "@fluentui/react";
import MESSAGES from "./locale-data/en.json";
import { DEFAULT_TEMPLATE_LOCALE } from "./resources";

export interface SystemConfig {
  authgearClientID: string;
  authgearEndpoint: string;
  sentryDSN: string;
  appHostSuffix: string;
  availableLanguages: string[];
  builtinLanguages: string[];
  themes: SystemConfigThemes;
  translations: SystemConfigTranslations;
  searchEnabled: boolean;
  web3Enabled: boolean;
  auditLogEnabled: boolean;
  gitCommitHash: string;
  analyticEnabled: boolean;
  analyticEpoch: string;
  gtmContainerID: string;
  uiImplementation: string;
  uiSettingsImplemenation: string;
}

export interface SystemConfigThemes {
  main: ITheme;
  inverted: ITheme;
  destructive: ITheme;
  actionButton: ITheme;
  verifyButton: ITheme;
  defaultButton: ITheme;
}

export interface SystemConfigTranslations {
  en: Record<string, string>;
}

export interface PartialSystemConfig
  extends Partial<Omit<SystemConfig, "themes" | "translations">> {
  themes?: Partial<Record<keyof SystemConfigThemes, IPartialTheme>>;
  translations?: Partial<SystemConfigTranslations>;
}

export const defaultSystemConfig: PartialSystemConfig = {
  themes: {
    // Generated with Fluent UI theme Designer with
    // Primary color: #176df3
    // Text color: #323130
    // Background color: #ffffff
    main: {
      palette: {
        themePrimary: "#176df3",
        themeLighterAlt: "#f5f9fe",
        themeLighter: "#d8e6fd",
        themeLight: "#b7d1fb",
        themeTertiary: "#70a4f7",
        themeSecondary: "#317bf4",
        themeDarkAlt: "#1460da",
        themeDark: "#1151b8",
        themeDarker: "#0c3c88",
        neutralLighterAlt: "#faf9f8",
        neutralLighter: "#f3f2f1",
        neutralLight: "#edebe9",
        neutralQuaternaryAlt: "#e1dfdd",
        neutralQuaternary: "#d0d0d0",
        neutralTertiaryAlt: "#c8c6c4",
        neutralTertiary: "#a19f9d",
        neutralSecondary: "#605e5c",
        neutralPrimaryAlt: "#3b3a39",
        neutralPrimary: "#323130",
        neutralDark: "#201f1e",
        black: "#000000",
        white: "#ffffff",
      },
    },
    // Generated with Fluent UI theme Designer with
    // Primary color: #ffffff
    // Text color: #ffffff
    // Background color: #176df3
    inverted: {
      palette: {
        themePrimary: "#ffffff",
        themeLighterAlt: "#767676",
        themeLighter: "#a6a6a6",
        themeLight: "#c8c8c8",
        themeTertiary: "#d0d0d0",
        themeSecondary: "#dadada",
        themeDarkAlt: "#eaeaea",
        themeDark: "#f4f4f4",
        themeDarker: "#f8f8f8",
        neutralLighterAlt: "#1567ec",
        neutralLighter: "#1566e8",
        neutralLight: "#1462de",
        neutralQuaternaryAlt: "#135bcf",
        neutralQuaternary: "#1257c6",
        neutralTertiaryAlt: "#1153be",
        neutralTertiary: "#c8c8c8",
        neutralSecondary: "#d0d0d0",
        neutralPrimaryAlt: "#dadada",
        neutralPrimary: "#ffffff",
        neutralDark: "#f4f4f4",
        black: "#f8f8f8",
        white: "#176df3",
      },
    },
    // Generated with Fluent UI theme Designer with
    // Primary color: #d81010
    // Text color: #d81010
    // Background color: #fbfcff
    destructive: {
      palette: {
        themePrimary: "#d81010",
        themeLighterAlt: "#fdf4f4",
        themeLighter: "#f9d4d4",
        themeLight: "#f4b0b0",
        themeTertiary: "#e86767",
        themeSecondary: "#dd2828",
        themeDarkAlt: "#c30e0e",
        themeDark: "#a50c0c",
        themeDarker: "#790808",
        neutralLighterAlt: "#f3f4f8",
        neutralLighter: "#eff0f4",
        neutralLight: "#e5e7ea",
        neutralQuaternaryAlt: "#d6d7da",
        neutralQuaternary: "#cccdd0",
        neutralTertiaryAlt: "#c4c5c8",
        neutralTertiary: "#f4b0b0",
        neutralSecondary: "#e86767",
        neutralPrimaryAlt: "#dd2828",
        neutralPrimary: "#d81010",
        neutralDark: "#a50c0c",
        black: "#790808",
        white: "#fbfcff",
      },
    },
    actionButton: {
      palette: {
        themePrimary: "#176df3",
        themeLighterAlt: "#f5f9fe",
        themeLighter: "#d8e6fd",
        themeLight: "#b7d1fb",
        themeTertiary: "#70a4f7",
        themeSecondary: "#317bf4",
        themeDarkAlt: "#1460da",
        themeDark: "#1151b8",
        themeDarker: "#0c3c88",
        neutralLighterAlt: "#faf9f8",
        neutralLighter: "#f3f2f1",
        neutralLight: "#edebe9",
        neutralQuaternaryAlt: "#e1dfdd",
        neutralQuaternary: "#d0d0d0",
        neutralTertiaryAlt: "#c8c6c4",
        neutralTertiary: "#b7d1fb",
        neutralSecondary: "#70a4f7",
        neutralPrimaryAlt: "#317bf4",
        neutralPrimary: "#176df3",
        neutralDark: "#1151b8",
        black: "#0c3c88",
        white: "#ffffff",
      },
    },
    // Generated with Fluent UI theme Designer with
    // Primary color: #10B070
    // Text color: #000000
    // Background color: #ffffff
    verifyButton: {
      palette: {
        themePrimary: "#10b070",
        themeLighterAlt: "#f3fcf8",
        themeLighter: "#cff2e4",
        themeLight: "#a8e7ce",
        themeTertiary: "#5ed0a2",
        themeSecondary: "#25b97e",
        themeDarkAlt: "#0e9e65",
        themeDark: "#0c8655",
        themeDarker: "#09633f",
        neutralLighterAlt: "#faf9f8",
        neutralLighter: "#f3f2f1",
        neutralLight: "#edebe9",
        neutralQuaternaryAlt: "#e1dfdd",
        neutralQuaternary: "#d0d0d0",
        neutralTertiaryAlt: "#c8c6c4",
        neutralTertiary: "#aaaaaa",
        neutralSecondary: "#373737",
        neutralPrimaryAlt: "#2f2f2f",
        neutralPrimary: "#000000",
        neutralDark: "#151515",
        black: "#0b0b0b",
        white: "#ffffff",
      },
    },
    defaultButton: {
      palette: {
        themePrimary: "#ffffff",
        themeLighterAlt: "#767676",
        themeLighter: "#a6a6a6",
        themeLight: "#c8c8c8",
        themeTertiary: "#d0d0d0",
        themeSecondary: "#dadada",
        themeDarkAlt: "#eaeaea",
        themeDark: "#f4f4f4",
        themeDarker: "#f8f8f8",
        neutralLighterAlt: "#3c3b39",
        neutralLighter: "#f3f2f1", // disable background
        neutralLight: "#514f4e",
        neutralQuaternaryAlt: "#595756",
        neutralQuaternary: "#5f5e5c",
        neutralTertiaryAlt: "#7a7977",
        neutralTertiary: "#aaaaaa", // disable text
        neutralSecondary: "#d0d0d0",
        neutralPrimaryAlt: "#dadada",
        neutralPrimary: "#ffffff",
        neutralDark: "#f4f4f4",
        black: "#f8f8f8",
        white: "#666666", // normal text
      },
    },
  },
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
    themes: {
      ...baseConfig.themes,
      ...overlayConfig.themes,
    },
    translations: {
      en: {
        ...baseConfig.translations?.en,
        ...overlayConfig.translations?.en,
      },
    },
  };
}

// eslint-disable-next-line complexity
export function instantiateSystemConfig(
  config: PartialSystemConfig
): SystemConfig {
  return {
    authgearClientID: config.authgearClientID ?? "",
    authgearEndpoint: config.authgearEndpoint ?? "",
    sentryDSN: config.sentryDSN ?? "",
    appHostSuffix: config.appHostSuffix ?? "",
    availableLanguages: config.availableLanguages ?? [DEFAULT_TEMPLATE_LOCALE],
    builtinLanguages: config.builtinLanguages ?? [DEFAULT_TEMPLATE_LOCALE],
    themes: {
      main: createTheme(config.themes?.main ?? {}),
      inverted: createTheme(config.themes?.inverted ?? {}),
      destructive: createTheme(config.themes?.destructive ?? {}),
      actionButton: createTheme(config.themes?.actionButton ?? {}),
      verifyButton: createTheme(config.themes?.verifyButton ?? {}),
      defaultButton: createTheme(config.themes?.defaultButton ?? {}),
    },
    translations: {
      en: config.translations?.en ?? {},
    },
    searchEnabled: config.searchEnabled ?? false,
    web3Enabled: config.web3Enabled ?? false,
    auditLogEnabled: config.auditLogEnabled ?? false,
    gitCommitHash: config.gitCommitHash ?? "",
    analyticEnabled: config.analyticEnabled ?? false,
    analyticEpoch: config.analyticEpoch ?? "",
    gtmContainerID: config.gtmContainerID ?? "",
    uiImplementation: config.uiImplementation ?? "interaction",
    uiSettingsImplemenation: config.uiSettingsImplemenation ?? "v1",
  };
}
