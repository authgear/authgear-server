import {
  AuthenticationFlowConfig,
  AuthenticationFlowStepConfig,
  AuthenticationFlowsConfig,
} from "../../../types";

export const flowTypes = [
  "signup_login",
  "signup",
  "login",
  "promote",
  "reauth",
  "account_recovery",
] as const;
export type FlowType = (typeof flowTypes)[number];

export function isFlowType(v: unknown): v is FlowType {
  return typeof v === "string" && (flowTypes as readonly string[]).includes(v);
}

export const DEFAULT_FLOW_NAME = "default";

const flowListKeys: Record<FlowType, keyof AuthenticationFlowsConfig> = {
  signup_login: "signup_login_flows",
  signup: "signup_flows",
  login: "login_flows",
  promote: "promote_flows",
  reauth: "reauth_flows",
  account_recovery: "account_recovery_flows",
};

export function flowListKey(type: FlowType): keyof AuthenticationFlowsConfig {
  return flowListKeys[type];
}

function flowsOf(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType
): AuthenticationFlowConfig[] {
  return config?.[flowListKeys[type]] ?? [];
}

// generatedAuthenticationFlows from the portal API: the generated flow of
// each type. Account recovery is null when the settings give no way to
// recover an account; promote is not generated.
export interface GeneratedFlows {
  signupFlow: AuthenticationFlowConfig;
  loginFlow: AuthenticationFlowConfig;
  signupLoginFlow: AuthenticationFlowConfig;
  reauthFlow: AuthenticationFlowConfig;
  accountRecoveryFlow?: AuthenticationFlowConfig | null;
}

// The generated flows shaped like authentication_flow, so they resolve the
// same way as custom ones.
export function generatedFlowsConfig(
  generated: GeneratedFlows
): AuthenticationFlowsConfig {
  return {
    signup_flows: [generated.signupFlow],
    login_flows: [generated.loginFlow],
    signup_login_flows: [generated.signupLoginFlow],
    reauth_flows: [generated.reauthFlow],
    ...(generated.accountRecoveryFlow != null
      ? { account_recovery_flows: [generated.accountRecoveryFlow] }
      : {}),
  };
}

export type FlowSource = "custom" | "generated";

export interface ResolvedFlow {
  flow: AuthenticationFlowConfig;
  source: FlowSource;
}

// Mirrors the runtime: a custom flow with the name wins, and only "default"
// falls back to the generated flow.
export function resolveFlow(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig,
  type: FlowType,
  name: string
): ResolvedFlow | null {
  const customFlow = flowsOf(custom, type).find((f) => f.name === name);
  if (customFlow != null) {
    return { flow: customFlow, source: "custom" };
  }
  if (name === DEFAULT_FLOW_NAME) {
    const generatedFlow = flowsOf(generated, type).find(
      (f) => f.name === DEFAULT_FLOW_NAME
    );
    if (generatedFlow != null) {
      return { flow: generatedFlow, source: "generated" };
    }
  }
  return null;
}

export function listFlowNames(
  custom: AuthenticationFlowsConfig | undefined,
  type: FlowType
): string[] {
  const names = [DEFAULT_FLOW_NAME];
  for (const flow of flowsOf(custom, type)) {
    if (!names.includes(flow.name)) {
      names.push(flow.name);
    }
  }
  return names;
}

// The editor's URL for a flow, or for a bundle when type is signup_login.
export function flowPath(appID: string, type: FlowType, name: string): string {
  return `/project/${encodeURIComponent(
    appID
  )}/configuration/authentication/flows/${type}/${encodeURIComponent(name)}`;
}

export function flowListPath(appID: string): string {
  return `/project/${encodeURIComponent(
    appID
  )}/configuration/authentication/flows`;
}

export interface AccountRecoveryDestination {
  channel: string;
  otp_form: string;
}

// What the server sends when select_destination lists no allowed_channels
// (config.GetAllAccountRecoveryChannel).
export const ALL_ACCOUNT_RECOVERY_DESTINATIONS: AccountRecoveryDestination[] = [
  { channel: "email", otp_form: "link" },
  { channel: "email", otp_form: "code" },
  { channel: "sms", otp_form: "code" },
  { channel: "sms", otp_form: "link" },
  { channel: "whatsapp", otp_form: "code" },
];

// What one select_destination step can send.
export function selectDestinationChannels(
  step: AuthenticationFlowStepConfig
): AccountRecoveryDestination[] {
  const allowed = step.allowed_channels ?? [];
  if (allowed.length === 0) {
    return ALL_ACCOUNT_RECOVERY_DESTINATIONS;
  }
  return allowed.flatMap((c) =>
    c.otp_form != null ? [{ channel: c.channel, otp_form: c.otp_form }] : []
  );
}

// Every channel and form (link, code) an account recovery flow can send,
// once each, from its select_destination steps. The verify step after them
// shows one section per destination, since a link and a code look different.
export function accountRecoveryDestinations(
  flow: AuthenticationFlowConfig
): AccountRecoveryDestination[] {
  const found = new Map<string, AccountRecoveryDestination>();
  const walk = (steps: AuthenticationFlowStepConfig[]) => {
    for (const step of steps) {
      if (step.type === "select_destination") {
        for (const d of selectDestinationChannels(step)) {
          const key = `${d.channel}.${d.otp_form}`;
          if (!found.has(key)) {
            found.set(key, d);
          }
        }
      }
      for (const option of step.one_of ?? []) {
        walk(option.steps ?? []);
      }
    }
  };
  walk(flow.steps ?? []);
  return [...found.values()];
}
