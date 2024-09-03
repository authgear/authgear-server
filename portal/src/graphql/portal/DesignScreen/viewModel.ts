import {
  CSSVariable,
  CssAstVisitor,
  CustomisableThemeStyleGroup,
  DEFAULT_DARK_THEME,
  DEFAULT_LIGHT_THEME,
  WatermarkDisabledDisplay,
  WatermarkEnabledDisplay,
  getThemeTargetSelector,
} from "../../../model/themeAuthFlowV2";
import { PortalAPIAppConfig } from "../../../types";
import { AppLogoResource, BranchDesignFormState, TranslationKey } from "./form";

export type PreviewPageType =
  | "Login"
  | "SignUp"
  | "EnterPassword"
  | "EnterOOBOTP"
  | "UsePasskey"
  | "EnterTOTP"
  | "OOBOTPLink"
  | "CreatePassword";

interface PreviewPageOption {
  key: PreviewPageType;
  screen: string;
}

const PreviewPage: Record<PreviewPageType, string> = {
  Login: "preview/login",
  SignUp: "preview/signup",
  EnterPassword: "preview/authflow/v2/enter_password",
  EnterOOBOTP: "preview/authflow/v2/enter_oob_otp",
  UsePasskey: "preview/authflow/v2/use_passkey",
  EnterTOTP: "preview/authflow/v2/enter_totp",
  OOBOTPLink: "preview/authflow/v2/oob_otp_link",
  CreatePassword: "preview/authflow/v2/create_password",
};

// eslint-disable-next-line complexity
export function getSupportedPreviewPagesFromConfig(
  config: PortalAPIAppConfig
): PreviewPageOption[] {
  const pages: PreviewPageOption[] = [];

  pages.push({ key: "Login", screen: PreviewPage.Login }); // Login page is always there
  pages.push({ key: "SignUp", screen: PreviewPage.SignUp }); // SignUp page is always there
 
  if (
    config.authentication?.primary_authenticators?.includes("oob_otp_sms") ||
    (config.authentication?.primary_authenticators?.includes("oob_otp_email") &&
      config.authenticator?.oob_otp?.email?.email_otp_mode === "code")
  ) {
    pages.push({ key: "EnterOOBOTP", screen: PreviewPage.EnterOOBOTP });
  }
  if (config.authentication?.primary_authenticators?.includes("password")) {
    pages.push({ key: "CreatePassword", screen: PreviewPage.CreatePassword });
    pages.push({ key: "EnterPassword", screen: PreviewPage.EnterPassword });
  }
  if (config.authentication?.primary_authenticators?.includes("passkey")) {
    pages.push({ key: "UsePasskey", screen: PreviewPage.UsePasskey });
  }
  if (
    config.authentication?.secondary_authentication_mode === "required" ||
    config.authentication?.secondary_authentication_mode === "if_exists"
  ) {
    pages.push({ key: "EnterTOTP", screen: PreviewPage.EnterTOTP });
  }
  if (
    config.authentication?.primary_authenticators?.includes("oob_otp_email") &&
    config.authenticator?.oob_otp?.email?.email_otp_mode === "login_link"
  ) {
    pages.push({ key: "OOBOTPLink", screen: PreviewPage.OOBOTPLink });
  }
  return pages;
}

export interface PreviewCustomisationMessage {
  type: "PreviewCustomisationMessage";
  theme: string;
  cssVars: Record<string, string>;
  images: Record<string, string | null>;
  translations: Record<string, string>;
}

export function mapDesignFormStateToPreviewCustomisationMessage(
  state: BranchDesignFormState
): PreviewCustomisationMessage {
  const cssAstVisitor = new CssAstVisitor(
    getThemeTargetSelector(state.selectedTheme)
  );
  const defaultStyleGroup = new CustomisableThemeStyleGroup(
    state.selectedTheme === "light" ? DEFAULT_LIGHT_THEME : DEFAULT_DARK_THEME
  );
  defaultStyleGroup.acceptCssAstVisitor(cssAstVisitor);
  const styleGroup = new CustomisableThemeStyleGroup(
    state.selectedTheme === "light"
      ? state.customisableLightTheme
      : {
          ...state.customisableDarkTheme,
          // fill common values that are only set in light theme
          card: {
            alignment: state.customisableLightTheme.card.alignment,
          },
          primaryButton: {
            ...state.customisableDarkTheme.primaryButton,
            borderRadius:
              state.customisableLightTheme.primaryButton.borderRadius,
          },
          secondaryButton: {
            borderRadius:
              state.customisableLightTheme.secondaryButton.borderRadius,
          },
          inputField: {
            borderRadius: state.customisableLightTheme.inputField.borderRadius,
          },
          link: {
            ...state.customisableDarkTheme.link,
            textDecoration: state.customisableLightTheme.link.textDecoration,
          },
        }
  );
  styleGroup.acceptCssAstVisitor(cssAstVisitor);
  const declarations = cssAstVisitor.getDeclarations();
  const cssVars: Record<string, string> = {};
  for (const declaration of declarations) {
    cssVars[declaration.prop] = declaration.value;
  }

  // Handle background image for both themes.
  if (state.selectedTheme === "light") {
    if (state.backgroundImageBase64EncodedData) {
      cssVars[
        CSSVariable.LayoutBackgroundImage
      ] = `url("data:;base64,${state.backgroundImageBase64EncodedData}")`;
    } else {
      cssVars[CSSVariable.LayoutBackgroundImage] = "initial";
    }
  } else {
    if (state.backgroundImageDarkBase64EncodedData) {
      cssVars[
        CSSVariable.LayoutBackgroundImage
      ] = `url("data:;base64,${state.backgroundImageDarkBase64EncodedData}")`;
    } else {
      cssVars[CSSVariable.LayoutBackgroundImage] = "initial";
    }
  }

  const theme = state.selectedTheme;

  cssVars[CSSVariable.WatermarkDisplay] = state.showAuthgearLogo
    ? WatermarkEnabledDisplay
    : WatermarkDisabledDisplay;

  const images: Record<string, string | null> = {};

  images["brand-logo-light"] = getLogoDataSrc(state.appLogo);
  images["brand-logo-dark"] = getLogoDataSrc(state.appLogoDark);

  const translations = {
    [TranslationKey.AppName]: state.appName,
  };

  return {
    type: "PreviewCustomisationMessage",
    theme,
    cssVars,
    images,
    translations,
  };
}

function getLogoDataSrc(logoResouce: AppLogoResource): string | null {
  const data =
    logoResouce.base64EncodedData ?? logoResouce.fallbackBase64EncodedData;
  if (!data) {
    return null;
  }
  return `data:;base64,${data}`;
}
