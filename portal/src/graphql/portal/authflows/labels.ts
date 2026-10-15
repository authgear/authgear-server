import {
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
  AuthenticatorPhoneOTPMode,
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

const phoneCodeLabelMessageIDs: Record<AuthenticatorPhoneOTPMode, string> = {
  sms: "AuthFlowsScreen.authentication.oob_otp_sms.sms",
  whatsapp: "AuthFlowsScreen.authentication.oob_otp_sms.whatsapp",
  whatsapp_sms: "AuthFlowsScreen.authentication.oob_otp_sms.whatsapp_sms",
};

// phoneOTPMode, from authenticator.oob_otp.sms.phone_otp_mode, names the
// channel a phone code is sent by.
export function optionLabel(
  option: AuthenticationFlowOptionConfig,
  renderToString: RenderToString,
  phoneOTPMode?: AuthenticatorPhoneOTPMode
): string {
  if (option.identification != null) {
    const id = identificationMessageIDs[option.identification];
    return id != null ? renderToString(id) : option.identification;
  }
  if (
    phoneOTPMode != null &&
    (option.authentication === "primary_oob_otp_sms" ||
      option.authentication === "secondary_oob_otp_sms")
  ) {
    return renderToString(phoneCodeLabelMessageIDs[phoneOTPMode]);
  }
  if (option.authentication != null) {
    const id = authenticationMessageIDs[option.authentication];
    return id != null ? renderToString(id) : option.authentication;
  }
  return renderToString("AuthFlowsScreen.option.unknown");
}

// Steps the runtime runs only when a condition holds, without the config
// marking them optional: a short condition shown on the step, and a note
// that explains it in full.
export interface IfNeededMessageIDs {
  condition: string;
  note: string;
}

const ifNeededMessageIDsByType: Partial<Record<string, IfNeededMessageIDs>> = {
  change_password: {
    condition: "AuthFlowsScreen.if-needed.condition.change_password",
    note: "AuthFlowsScreen.if-needed.change_password",
  },
  terminate_other_sessions: {
    condition: "AuthFlowsScreen.if-needed.condition.terminate_other_sessions",
    note: "AuthFlowsScreen.if-needed.terminate_other_sessions",
  },
  prompt_create_passkey: {
    condition: "AuthFlowsScreen.if-needed.condition.prompt_create_passkey",
    note: "AuthFlowsScreen.if-needed.prompt_create_passkey",
  },
};

export function ifNeededMessageIDs(type: string): IfNeededMessageIDs | null {
  return ifNeededMessageIDsByType[type] ?? null;
}

// Authentications that only ever serve as a second factor.
const secondFactorAuthentications = new Set([
  "secondary_password",
  "secondary_totp",
  "secondary_oob_otp_email",
  "secondary_oob_otp_sms",
  "recovery_code",
  "device_token",
]);

// An optional step is skipped when the user has none of its options set up.
// When every option is a second factor, that is "has set up 2FA".
export function optionalMessageIDs(
  step: AuthenticationFlowStepConfig
): IfNeededMessageIDs {
  const options = step.one_of ?? [];
  const is2FA =
    options.length > 0 &&
    options.every(
      (o) =>
        o.authentication != null &&
        secondFactorAuthentications.has(o.authentication)
    );
  return is2FA
    ? {
        condition: "AuthFlowsScreen.optional.condition.2fa",
        note: "AuthFlowsScreen.optional.note.2fa",
      }
    : {
        condition: "AuthFlowsScreen.optional.condition.any",
        note: "AuthFlowsScreen.optional.note.any",
      };
}

// What each step does, shown in the side panel.
const stepDescriptionMessageIDs: Partial<Record<string, string>> = {
  identify: "AuthFlowsScreen.step-description.identify",
  authenticate: "AuthFlowsScreen.step-description.authenticate",
  create_authenticator: "AuthFlowsScreen.step-description.create_authenticator",
  verify: "AuthFlowsScreen.step-description.verify",
  fill_in_user_profile: "AuthFlowsScreen.step-description.fill_in_user_profile",
  fill_form: "AuthFlowsScreen.step-description.fill_form",
  view_recovery_code: "AuthFlowsScreen.step-description.view_recovery_code",
  prompt_create_passkey:
    "AuthFlowsScreen.step-description.prompt_create_passkey",
  terminate_other_sessions:
    "AuthFlowsScreen.step-description.terminate_other_sessions",
  check_account_status: "AuthFlowsScreen.step-description.check_account_status",
  change_password: "AuthFlowsScreen.step-description.change_password",
  select_destination: "AuthFlowsScreen.step-description.select_destination",
  verify_account_recovery_code:
    "AuthFlowsScreen.step-description.verify_account_recovery_code",
  reset_password: "AuthFlowsScreen.step-description.reset_password",
};

export function stepDescriptionMessageID(type: string): string | null {
  return stepDescriptionMessageIDs[type] ?? null;
}

const phoneCodeSetupMessageIDs: Record<AuthenticatorPhoneOTPMode, string> = {
  sms: "AuthFlowsScreen.setup.oob_otp_sms.sms",
  whatsapp: "AuthFlowsScreen.setup.oob_otp_sms.whatsapp",
  whatsapp_sms: "AuthFlowsScreen.setup.oob_otp_sms.whatsapp_sms",
};

// Phone codes go by SMS, WhatsApp or either, per
// authenticator.oob_otp.sms.phone_otp_mode.
export function phoneCodeSetupMessageID(
  mode: AuthenticatorPhoneOTPMode | undefined
): string {
  return mode != null
    ? phoneCodeSetupMessageIDs[mode]
    : "AuthFlowsScreen.setup.oob_otp_sms.any";
}

const verifyPhoneMessageIDs: Record<AuthenticatorPhoneOTPMode, string> = {
  sms: "AuthFlowsScreen.verify.phone.sms",
  whatsapp: "AuthFlowsScreen.verify.phone.whatsapp",
  whatsapp_sms: "AuthFlowsScreen.verify.phone.whatsapp_sms",
};

// What a verify step asks for, named by the login ID its branch collected.
export function verifyCodeMessageID(
  identification: string | null,
  phoneOTPMode: AuthenticatorPhoneOTPMode | undefined
): string {
  if (identification === "email") {
    return "AuthFlowsScreen.verify.email";
  }
  if (identification === "phone" && phoneOTPMode != null) {
    return verifyPhoneMessageIDs[phoneOTPMode];
  }
  return "AuthFlowsScreen.sketch.enter-code";
}

// A pageless step the user never sees can still need saying what it does:
// a short line on the pill, and a note that explains it.
export function pagelessNoteMessageIDs(
  step: AuthenticationFlowStepConfig
): { line: string; note: string } | null {
  const options = step.one_of ?? [];
  if (
    step.type === "identify" &&
    options.length > 0 &&
    options.every((o) => o.identification === "id_token")
  ) {
    return {
      line: "AuthFlowsScreen.pageless.id_token",
      note: "AuthFlowsScreen.pageless.id_token.note",
    };
  }
  return null;
}

const destinationMessageIDs: Partial<Record<string, string>> = {
  "email.link": "AuthFlowsScreen.destination.email.link",
  "email.code": "AuthFlowsScreen.destination.email.code",
  "sms.link": "AuthFlowsScreen.destination.sms.link",
  "sms.code": "AuthFlowsScreen.destination.sms.code",
  "whatsapp.link": "AuthFlowsScreen.destination.whatsapp.link",
  "whatsapp.code": "AuthFlowsScreen.destination.whatsapp.code",
};

// What an account recovery channel sends, and where: a link or a code
// (otp_form), by email, SMS or WhatsApp, to the login ID the user entered.
export function destinationMessageID(
  channel: string,
  otpForm: string | undefined
): string | null {
  return destinationMessageIDs[`${channel}.${otpForm}`] ?? null;
}
