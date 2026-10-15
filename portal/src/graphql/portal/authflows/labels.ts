import {
  AuthenticationFlowOptionConfig,
  OAuthSSOProviderType,
} from "../../../types";
import { IntlContextValue } from "../../../intl";
import { FlowType } from "./model";

type RenderToString = IntlContextValue["renderToString"];

export const flowTypeMessageIDs: Record<FlowType, string> = {
  signup_login: "AuthFlowsScreen.flow-type.signup_login",
  signup: "AuthFlowsScreen.flow-type.signup",
  login: "AuthFlowsScreen.flow-type.login",
  promote: "AuthFlowsScreen.flow-type.promote",
  reauth: "AuthFlowsScreen.flow-type.reauth",
  account_recovery: "AuthFlowsScreen.flow-type.account_recovery",
};

export const oauthProviderMessageIDs: Record<OAuthSSOProviderType, string> = {
  apple: "oauth-provider.apple",
  google: "oauth-provider.google",
  facebook: "oauth-provider.facebook",
  github: "oauth-provider.github",
  linkedin: "oauth-provider.linkedin",
  azureadv2: "oauth-provider.azureadv2",
  azureadb2c: "oauth-provider.azureadb2c",
  adfs: "oauth-provider.adfs",
  wechat: "oauth-provider.wechat",
};

const stepTitleMessageIDs: Partial<Record<string, string>> = {
  identify: "AuthFlowsScreen.step.identify",
  authenticate: "AuthFlowsScreen.step.authenticate",
  create_authenticator: "AuthFlowsScreen.step.create_authenticator",
  verify: "AuthFlowsScreen.step.verify",
  fill_in_user_profile: "AuthFlowsScreen.step.fill_in_user_profile",
  fill_form: "AuthFlowsScreen.step.fill_form",
  view_recovery_code: "AuthFlowsScreen.step.view_recovery_code",
  prompt_create_passkey: "AuthFlowsScreen.step.prompt_create_passkey",
  terminate_other_sessions: "AuthFlowsScreen.step.terminate_other_sessions",
  check_account_status: "AuthFlowsScreen.step.check_account_status",
  change_password: "AuthFlowsScreen.step.change_password",
  select_destination: "AuthFlowsScreen.step.select_destination",
  verify_account_recovery_code:
    "AuthFlowsScreen.step.verify_account_recovery_code",
  reset_password: "AuthFlowsScreen.step.reset_password",
};

const identificationMessageIDs: Partial<Record<string, string>> = {
  email: "AuthFlowsScreen.identification.email",
  phone: "AuthFlowsScreen.identification.phone",
  username: "AuthFlowsScreen.identification.username",
  oauth: "AuthFlowsScreen.identification.oauth",
  passkey: "AuthFlowsScreen.identification.passkey",
  select_account: "AuthFlowsScreen.identification.select_account",
  id_token: "AuthFlowsScreen.identification.id_token",
  ldap: "AuthFlowsScreen.identification.ldap",
};

const authenticationMessageIDs: Partial<Record<string, string>> = {
  primary_password: "AuthFlowsScreen.authentication.password",
  secondary_password: "AuthFlowsScreen.authentication.password",
  primary_passkey: "AuthFlowsScreen.authentication.passkey",
  primary_oob_otp_email: "AuthFlowsScreen.authentication.oob_otp_email",
  secondary_oob_otp_email: "AuthFlowsScreen.authentication.oob_otp_email",
  primary_oob_otp_sms: "AuthFlowsScreen.authentication.oob_otp_sms",
  secondary_oob_otp_sms: "AuthFlowsScreen.authentication.oob_otp_sms",
  secondary_totp: "AuthFlowsScreen.authentication.totp",
  recovery_code: "AuthFlowsScreen.authentication.recovery_code",
  device_token: "AuthFlowsScreen.authentication.device_token",
};

// Unknown step types fall back to the raw type, so newer configs still render.
export function stepTitle(
  type: string,
  renderToString: RenderToString
): string {
  const id = stepTitleMessageIDs[type];
  return id != null ? renderToString(id) : type;
}

export function optionLabel(
  option: AuthenticationFlowOptionConfig,
  renderToString: RenderToString
): string {
  if (option.identification != null) {
    const id = identificationMessageIDs[option.identification];
    return id != null ? renderToString(id) : option.identification;
  }
  if (option.authentication != null) {
    const id = authenticationMessageIDs[option.authentication];
    return id != null ? renderToString(id) : option.authentication;
  }
  return renderToString("AuthFlowsScreen.option.unknown");
}
