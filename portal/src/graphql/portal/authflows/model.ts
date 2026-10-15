import {
  AuthenticationFlowConfig,
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
export const DEFAULT_FLOW_TYPE: FlowType = "signup_login";

const flowListKeys: Record<FlowType, keyof AuthenticationFlowsConfig> = {
  signup_login: "signup_login_flows",
  signup: "signup_flows",
  login: "login_flows",
  promote: "promote_flows",
  reauth: "reauth_flows",
  account_recovery: "account_recovery_flows",
};

function flowsOf(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType
): AuthenticationFlowConfig[] {
  return config?.[flowListKeys[type]] ?? [];
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

export interface FlowSelection {
  type: FlowType;
  name: string;
}

export function parseFlowSelection(
  params: URLSearchParams,
  custom: AuthenticationFlowsConfig | undefined
): FlowSelection {
  const rawType = params.get("type");
  const type = isFlowType(rawType) ? rawType : DEFAULT_FLOW_TYPE;
  const rawName = params.get("name");
  const name =
    rawName != null && listFlowNames(custom, type).includes(rawName)
      ? rawName
      : DEFAULT_FLOW_NAME;
  return { type, name };
}

export function flowSelectionSearch(selection: FlowSelection): string {
  return `?${new URLSearchParams({
    type: selection.type,
    name: selection.name,
  }).toString()}`;
}
