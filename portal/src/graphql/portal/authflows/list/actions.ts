import { produce } from "immer";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";

export type FlowAction =
  // displayName, when given, is set too; empty removes it.
  | {
      kind: "rename";
      type: FlowType;
      name: string;
      newName: string;
      displayName?: string;
    }
  | { kind: "add"; type: FlowType; flow: AuthenticationFlowConfig }
  | { kind: "delete"; type: FlowType; name: string };

export const NAME_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

// The signup_login option field that names a flow of this type, if any.
function handoffKey(type: FlowType): "login_flow" | "signup_flow" | null {
  if (type === "login") {
    return "login_flow";
  }
  if (type === "signup") {
    return "signup_flow";
  }
  return null;
}

export function applyFlowAction(
  flows: AuthenticationFlowsConfig | undefined,
  action: FlowAction
): AuthenticationFlowsConfig {
  return produce(flows ?? {}, (d) => {
    const key = flowListKey(action.type);
    switch (action.kind) {
      case "add":
        (d[key] ??= []).push(action.flow);
        return;
      case "delete": {
        const rest = (d[key] ?? []).filter((f) => f.name !== action.name);
        // A flow list must have at least one flow, or be absent.
        if (rest.length > 0) {
          d[key] = rest;
        } else {
          delete d[key];
        }
        return;
      }
      case "rename": {
        const flow = (d[key] ?? []).find((f) => f.name === action.name);
        if (flow == null) {
          return;
        }
        if (action.displayName != null) {
          if (action.displayName === "") {
            delete flow.display_name;
          } else {
            flow.display_name = action.displayName;
          }
        }
        if (action.newName === action.name) {
          return;
        }
        flow.name = action.newName;
        // signup_login hand-offs that named the flow follow it.
        const field = handoffKey(action.type);
        if (field == null) {
          return;
        }
        for (const f of d.signup_login_flows ?? []) {
          for (const step of f.steps ?? []) {
            for (const o of step.one_of ?? []) {
              if (o[field] === action.name) {
                o[field] = action.newName;
              }
            }
          }
        }
      }
    }
  });
}

// The signup_login flows that still hand off to this flow. Removing a
// default flow resets it: the generated default takes its place under the
// same name, so hand-offs to it still work.
export function deleteBlockers(
  flows: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  name: string
): string[] {
  const field = handoffKey(type);
  if (field == null || name === DEFAULT_FLOW_NAME) {
    return [];
  }
  return (flows?.signup_login_flows ?? [])
    .filter((f) =>
      (f.steps ?? []).some((s) =>
        (s.one_of ?? []).some((o) => o[field] === name)
      )
    )
    .map((f) => f.name);
}

export type NameError = "invalid" | "taken" | "reserved";

// default is reserved: it is the generated flow, or the custom flow saved
// from editing it, never a name picked for a new or renamed flow.
export function nameError(
  flows: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  name: string
): NameError | null {
  if (!NAME_PATTERN.test(name)) {
    return "invalid";
  }
  if (name === DEFAULT_FLOW_NAME) {
    return "reserved";
  }
  if ((flows?.[flowListKey(type)] ?? []).some((f) => f.name === name)) {
    return "taken";
  }
  return null;
}

// The first of base_2, base_3, … that no flow of the type has.
// A flow name made from a display name: lowercase words joined by
// underscores, not starting with a digit. Empty when nothing is usable.
export function apiNameFrom(displayName: string): string {
  const name = displayName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "");
  return /^[0-9]/.test(name) ? `_${name}` : name;
}

export function freeName(
  flows: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  base: string
): string {
  for (let n = 2; ; n++) {
    const name = `${base}_${n}`;
    if (nameError(flows, type, name) == null) {
      return name;
    }
  }
}
