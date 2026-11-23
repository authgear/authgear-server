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

function flowsOf(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType
): AuthenticationFlowConfig[] {
  return config?.[flowListKey(type)] ?? [];
}

export interface FlowRow {
  type: FlowType;
  name: string;
  displayName?: string;
  // Named default: the flow AuthUI runs when the app names none.
  isDefault: boolean;
  // Generated from Login methods: the type has no custom default.
  generated: boolean;
}

// Only the creatable types are listed for now; other flows stay in
// authgear.yaml, untouched.
export type SectionKey = CreatableType;

export interface FlowSection {
  key: SectionKey;
  rows: FlowRow[];
}

// A type with no custom default runs the generated one. default comes first.
function rowsOf(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig,
  type: FlowType
): FlowRow[] {
  const customs = flowsOf(custom, type);
  const rows: FlowRow[] = [];
  if (
    !customs.some((f) => f.name === DEFAULT_FLOW_NAME) &&
    flowsOf(generated, type).some((f) => f.name === DEFAULT_FLOW_NAME)
  ) {
    rows.push({
      type,
      name: DEFAULT_FLOW_NAME,
      isDefault: true,
      generated: true,
    });
  }
  for (const { name, display_name } of customs) {
    rows.push({
      type,
      name,
      displayName: display_name,
      isDefault: name === DEFAULT_FLOW_NAME,
      generated: false,
    });
  }
  // default leads its section, wherever authgear.yaml has it.
  return [
    ...rows.filter((r) => r.isDefault),
    ...rows.filter((r) => !r.isDefault),
  ];
}

// Sections with no flows are left out.
export function flowSections(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig
): FlowSection[] {
  return CREATABLE_TYPES.map((key) => ({
    key,
    rows: rowsOf(custom, generated, key),
  })).filter((s) => s.rows.length > 0);
}
