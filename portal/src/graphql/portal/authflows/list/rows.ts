import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";

export type CreatableType = "signup_login" | "signup" | "login";
export const CREATABLE_TYPES: readonly CreatableType[] = [
  "signup_login",
  "signup",
  "login",
];

const BUNDLE_TYPES: readonly FlowType[] = ["signup_login", "signup", "login"];
export function bundleTypes(): readonly FlowType[] {
  return BUNDLE_TYPES;
}

const OTHER_TYPES: readonly FlowType[] = [
  "promote",
  "reauth",
  "account_recovery",
];

function flowsOf(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType
): AuthenticationFlowConfig[] {
  return config?.[flowListKey(type)] ?? [];
}

function has(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  name: string
): boolean {
  return flowsOf(config, type).some((f) => f.name === name);
}

// Whether the signup_login flow `name` only hands off to the signup and
// login flows of the same name, and both exist.
function bundles(
  flow: AuthenticationFlowConfig,
  hasMember: (type: FlowType) => boolean
): boolean {
  const options = (flow.steps ?? []).flatMap((s) => s.one_of ?? []);
  const handoffs = options.filter(
    (o) => o.login_flow != null || o.signup_flow != null
  );
  return (
    handoffs.length > 0 &&
    options.every(
      (o) =>
        (o.login_flow == null || o.login_flow === flow.name) &&
        (o.signup_flow == null || o.signup_flow === flow.name)
    ) &&
    hasMember("signup") &&
    hasMember("login")
  );
}

export function isBundle(
  custom: AuthenticationFlowsConfig | undefined,
  name: string,
  generated?: AuthenticationFlowsConfig
): boolean {
  const lookup = (type: FlowType) =>
    flowsOf(custom, type).find((f) => f.name === name) ??
    (name === DEFAULT_FLOW_NAME
      ? flowsOf(generated, type).find((f) => f.name === name)
      : undefined);
  const flow = lookup("signup_login");
  return flow != null && bundles(flow, (t) => lookup(t) != null);
}

export interface FlowRow {
  type: FlowType;
  name: string;
  generated: boolean;
  inUse: boolean;
  bundle: boolean;
  partOf: string | null;
}

export type SectionKey = "signup_login" | "signup" | "login" | "other";

export interface FlowSection {
  key: SectionKey;
  rows: FlowRow[];
}

function rowsOf(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig,
  type: FlowType
): FlowRow[] {
  const names: { name: string; generated: boolean }[] = [];
  if (
    !has(custom, type, DEFAULT_FLOW_NAME) &&
    has(generated, type, DEFAULT_FLOW_NAME)
  ) {
    names.push({ name: DEFAULT_FLOW_NAME, generated: true });
  }
  for (const f of flowsOf(custom, type)) {
    names.push({ name: f.name, generated: false });
  }
  return names.map(({ name, generated: g }) => {
    const bundled = isBundle(custom, name, generated);
    return {
      type,
      name,
      generated: g,
      inUse: name === DEFAULT_FLOW_NAME,
      bundle: type === "signup_login" && bundled,
      partOf: (type === "signup" || type === "login") && bundled ? name : null,
    };
  });
}

export function flowSections(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig
): FlowSection[] {
  const sections: FlowSection[] = (
    ["signup_login", "signup", "login"] as const
  ).map((key) => ({ key, rows: rowsOf(custom, generated, key) }));
  const other = OTHER_TYPES.flatMap((t) => rowsOf(custom, generated, t));
  if (other.length > 0) {
    sections.push({ key: "other", rows: other });
  }
  return sections;
}
