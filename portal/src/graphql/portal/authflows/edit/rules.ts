import {
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType } from "../model";

// What the authentication_flow schema (pkg/lib/config/authentication_flow.go)
// allows per flow type, in schema order. The first value is the default for a
// new step or option. fill_in_user_profile is deprecated, so it is not offered.

const signupStepTypes = [
  "identify",
  "create_authenticator",
  "verify",
  "fill_form",
  "view_recovery_code",
  "prompt_create_passkey",
];

const stepTypes: Record<FlowType, string[]> = {
  signup: signupStepTypes,
  promote: signupStepTypes,
  login: [
    "identify",
    "authenticate",
    "check_account_status",
    "terminate_other_sessions",
    "change_password",
    "prompt_create_passkey",
    "fill_form",
  ],
  signup_login: ["identify"],
  reauth: ["identify", "authenticate"],
  account_recovery: [
    "identify",
    "select_destination",
    "verify_account_recovery_code",
    "reset_password",
  ],
};

const signupIdentifications = [
  "email",
  "phone",
  "username",
  "oauth",
  "passkey",
  "ldap",
];
// id_token is left out of login and signup_login: apps practically never
// send one there. An existing id_token option still shows its value.
const loginIdentifications = [...signupIdentifications, "select_account"];

const identifications: Record<FlowType, string[]> = {
  signup: signupIdentifications,
  promote: signupIdentifications,
  login: loginIdentifications,
  signup_login: loginIdentifications,
  reauth: ["id_token"],
  account_recovery: ["email", "phone", "username"],
};

const createAuthentications = [
  "primary_password",
  "primary_oob_otp_email",
  "primary_oob_otp_sms",
  "secondary_password",
  "secondary_totp",
  "secondary_oob_otp_email",
  "secondary_oob_otp_sms",
];
const reauthAuthentications = [
  "primary_password",
  "primary_passkey",
  "primary_oob_otp_email",
  "primary_oob_otp_sms",
  "secondary_password",
  "secondary_totp",
  "secondary_oob_otp_email",
  "secondary_oob_otp_sms",
];
const loginAuthentications = [
  ...reauthAuthentications,
  "recovery_code",
  "device_token",
];

// Primary authenticators sign the user in; the rest are 2FA: a second
// factor, a recovery code, or a remembered device that skips 2FA.
export type AuthenticationKind = "primary" | "secondary";

export function authenticationKind(authentication: string): AuthenticationKind {
  return authentication.startsWith("primary_") ? "primary" : "secondary";
}

const codeAuthentications = new Set([
  "primary_oob_otp_email",
  "primary_oob_otp_sms",
  "secondary_oob_otp_email",
  "secondary_oob_otp_sms",
]);

// Authenticators that send codes to an email or phone. Only these read an
// option's target_step (the step whose email or phone they send to) and
// verification_required.
export function sendsCodes(authentication: string | undefined): boolean {
  return authentication != null && codeAuthentications.has(authentication);
}

export const CHOICE_STEP_TYPES: readonly string[] = [
  "identify",
  "authenticate",
  "create_authenticator",
];

// The step type a step's target_step must name.
export const TARGET_STEP_TYPES: Partial<Record<string, string>> = {
  verify: "identify",
  change_password: "authenticate",
};

export const IDENTIFICATION_SELECT_ACCOUNT = "select_account";
const IDENTIFICATION_PASSKEY = "passkey";

// signup_login options that always find an account, so they only continue to
// the login flow: the schema forbids signup_flow on select_account, and the
// runtime never reads it on passkey.
export function isLoginOnlyIdentification(
  identification: string | undefined
): boolean {
  return (
    identification === IDENTIFICATION_SELECT_ACCOUNT ||
    identification === IDENTIFICATION_PASSKEY
  );
}

export function allowedStepTypes(flowType: FlowType): string[] {
  return stepTypes[flowType];
}

export function allowedIdentifications(flowType: FlowType): string[] {
  return identifications[flowType];
}

export function allowedAuthentications(
  flowType: FlowType,
  stepType: string
): string[] {
  if (stepType === "create_authenticator") {
    return createAuthentications;
  }
  if (stepType === "authenticate") {
    return flowType === "login" ? loginAuthentications : reauthAuthentications;
  }
  return [];
}

function identifyOption(
  flowType: FlowType,
  identification: string
): AuthenticationFlowOptionConfig {
  if (flowType === "signup_login") {
    return isLoginOnlyIdentification(identification)
      ? { identification, login_flow: DEFAULT_FLOW_NAME }
      : {
          identification,
          login_flow: DEFAULT_FLOW_NAME,
          signup_flow: DEFAULT_FLOW_NAME,
        };
  }
  if (flowType === "account_recovery") {
    // Account recovery options need at least one nested step.
    return { identification, steps: [{ type: "select_destination" }] };
  }
  return { identification };
}

function optionWith(
  flowType: FlowType,
  stepType: string,
  value: string
): AuthenticationFlowOptionConfig {
  return stepType === "identify"
    ? identifyOption(flowType, value)
    : { authentication: value };
}

function optionValues(flowType: FlowType, stepType: string): string[] {
  return stepType === "identify"
    ? allowedIdentifications(flowType)
    : allowedAuthentications(flowType, stepType);
}

// A step as inserted: valid for the schema. targetStep is the name a verify
// or change_password step targets; formPointer, the attribute a new form
// asks for first.
export function newStep(
  flowType: FlowType,
  type: string,
  targetStep: string | null,
  formPointer: string | null = null
): AuthenticationFlowStepConfig {
  if (CHOICE_STEP_TYPES.includes(type)) {
    const [first] = optionValues(flowType, type);
    return { type, one_of: [optionWith(flowType, type, first)] };
  }
  if (TARGET_STEP_TYPES[type] != null && targetStep != null) {
    return { type, target_step: targetStep };
  }
  if (type === "fill_form" && formPointer != null) {
    return {
      type,
      form: { fields: [{ user_profile: { pointer: formPointer } }] },
    };
  }
  return { type };
}

// The next option for a choice step: the first value it does not list yet.
export function newOption(
  flowType: FlowType,
  step: AuthenticationFlowStepConfig
): AuthenticationFlowOptionConfig | null {
  if (!CHOICE_STEP_TYPES.includes(step.type)) {
    return null;
  }
  const used = new Set(
    (step.one_of ?? []).map((o) =>
      step.type === "identify" ? o.identification : o.authentication
    )
  );
  const value = optionValues(flowType, step.type).find((v) => !used.has(v));
  return value != null ? optionWith(flowType, step.type, value) : null;
}
