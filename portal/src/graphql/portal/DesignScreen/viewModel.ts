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
import { BranchDesignFormState, TranslationKey } from "./form";

export enum PreviewPage {
  Login = "preview/login",
  SignUp = "preview/signup",
  EnterPassword = "preview/authflow/v2/enter_password",
  EnterOOBOTP = "preview/authflow/v2/enter_oob_otp",
  UsePasskey = "preview/authflow/v2/use_passkey",
  EnterTOTP = "preview/authflow/v2/enter_totp",
  OOBOTPLink = "preview/authflow/v2/oob_otp_link",
  CreatePassword = "preview/authflow/v2/create_password",
}

export function getSupportedPreviewPagesFromConfig(
  config: PortalAPIAppConfig
): PreviewPage[] {
  const pages: PreviewPage[] = [];
  pages.push(PreviewPage.Login); // Login page is always there
  pages.push(PreviewPage.SignUp); // SignUp page is always there
  if (
    config.authentication?.primary_authenticators?.includes("oob_otp_sms") ||
    (config.authentication?.primary_authenticators?.includes("oob_otp_email") &&
      config.authenticator?.oob_otp?.email?.email_otp_mode === "code")
  ) {
    pages.push(PreviewPage.EnterOOBOTP);
  }
  if (config.authentication?.primary_authenticators?.includes("password")) {
    pages.push(PreviewPage.CreatePassword);
    pages.push(PreviewPage.EnterPassword);
  }
  if (config.authentication?.primary_authenticators?.includes("passkey")) {
    pages.push(PreviewPage.UsePasskey);
  }
  if (
    config.authentication?.secondary_authentication_mode === "required" ||
    config.authentication?.secondary_authentication_mode === "if_exists"
  ) {
    pages.push(PreviewPage.EnterTOTP);
  }
  if (
    config.authentication?.primary_authenticators?.includes("oob_otp_email") &&
    config.authenticator?.oob_otp?.email?.email_otp_mode === "login_link"
  ) {
    pages.push(PreviewPage.OOBOTPLink);
  }
  return pages;
}

export interface PreviewCustomisationMessage {
  type: "PreviewCustomisationMessage";
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
      : state.customisableDarkTheme
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

  cssVars[CSSVariable.WatermarkDisplay] = state.showAuthgearLogo
    ? WatermarkEnabledDisplay
    : WatermarkDisabledDisplay;

  const images: Record<string, string | null> = {};

  images["brand-logo-light"] = state.appLogoBase64EncodedData
    ? `data:;base64,${state.appLogoBase64EncodedData}`
    : null;
  images["brand-logo-dark"] = state.appLogoDarkBase64EncodedData
    ? `data:;base64,${state.appLogoDarkBase64EncodedData}`
    : null;

  const translations = {
    [TranslationKey.AppName]: state.appName,
  };

  return {
    type: "PreviewCustomisationMessage",
    cssVars,
    images,
    translations,
  };
}
