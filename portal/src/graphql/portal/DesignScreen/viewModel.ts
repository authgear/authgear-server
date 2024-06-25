import { PortalAPIAppConfig } from "../../../types";

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
