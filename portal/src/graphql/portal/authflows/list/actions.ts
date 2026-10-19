import { produce } from "immer";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";
import { bundleTypes } from "./rows";

export type FlowAction =
  | { kind: "set_default"; type: FlowType; name: string; bundle: boolean }
  | {
      kind: "rename";
      type: FlowType;
      name: string;
      newName: string;
      bundle: boolean;
    }
  | { kind: "add"; flows: { type: FlowType; flow: AuthenticationFlowConfig }[] }
  | { kind: "delete"; type: FlowType; name: string; bundle: boolean };

const NAME_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

function typesOf(type: FlowType, bundle: boolean): readonly FlowType[] {
  return bundle ? bundleTypes() : [type];
}

// Renames per type, as old name → new name, then points every signup_login
// hand-off at the new names.
function renameAll(
  d: AuthenticationFlowsConfig,
  renames: { type: FlowType; from: string; to: string }[]
) {
  // Only flows written in the config are renamed. A hand-off to a name with
  // no custom flow (a generated default) keeps pointing at that name.
  const done = renames.filter(({ type, from, to }) => {
    let renamed = false;
    for (const f of d[flowListKey(type)] ?? []) {
      if (f.name === from) {
        f.name = to;
        renamed = true;
      }
    }
    return renamed;
  });
  const map = (type: "login" | "signup", v: string | undefined) => {
    const r = done.find((x) => x.type === type && x.from === v);
    return r != null ? r.to : v;
  };
  for (const f of d.signup_login_flows ?? []) {
    for (const step of f.steps ?? []) {
      for (const o of step.one_of ?? []) {
        if (o.login_flow != null) o.login_flow = map("login", o.login_flow);
        if (o.signup_flow != null) o.signup_flow = map("signup", o.signup_flow);
      }
    }
  }
}

export function copyAs(
  sources: { type: FlowType; flow: AuthenticationFlowConfig }[],
  newName: string
): { type: FlowType; flow: AuthenticationFlowConfig }[] {
  return sources.map(({ type, flow }) => {
    const copy = produce(flow, (d) => {
      const old = d.name;
      d.name = newName;
      for (const step of d.steps ?? []) {
        for (const o of step.one_of ?? []) {
          if (o.login_flow === old) o.login_flow = newName;
          if (o.signup_flow === old) o.signup_flow = newName;
        }
      }
    });
    return { type, flow: copy };
  });
}

export function applyFlowAction(
  flows: AuthenticationFlowsConfig | undefined,
  action: FlowAction
): AuthenticationFlowsConfig {
  return produce(flows ?? {}, (d) => {
    switch (action.kind) {
      case "add":
        for (const { type, flow } of action.flows) {
          (d[flowListKey(type)] ??= []).push(flow);
        }
        return;
      case "delete":
        for (const t of typesOf(action.type, action.bundle)) {
          d[flowListKey(t)] = (d[flowListKey(t)] ?? []).filter(
            (f) => f.name !== action.name
          );
        }
        return;
      case "rename":
        renameAll(
          d,
          typesOf(action.type, action.bundle).map((t) => ({
            type: t,
            from: action.name,
            to: action.newName,
          }))
        );
        return;
      case "set_default": {
        // A placeholder no flow can be named keeps the swap from colliding.
        const tmp = " swapping";
        const types = typesOf(action.type, action.bundle);
        renameAll(
          d,
          types.map((t) => ({ type: t, from: DEFAULT_FLOW_NAME, to: tmp }))
        );
        renameAll(
          d,
          types.map((t) => ({
            type: t,
            from: action.name,
            to: DEFAULT_FLOW_NAME,
          }))
        );
        renameAll(
          d,
          types.map((t) => ({ type: t, from: tmp, to: action.name }))
        );
      }
    }
  });
}

export function deleteBlockers(
  flows: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  name: string,
  bundle: boolean
): string[] {
  if (type !== "signup" && type !== "login" && !bundle) {
    return [];
  }
  const key = type === "login" ? "login_flow" : "signup_flow";
  return (flows?.signup_login_flows ?? [])
    .filter((f) => !(bundle && f.name === name))
    .filter((f) =>
      (f.steps ?? []).some((s) =>
        (s.one_of ?? []).some((o) =>
          bundle
            ? o.login_flow === name || o.signup_flow === name
            : o[key] === name
        )
      )
    )
    .map((f) => f.name);
}

export type NameError = "invalid" | "taken" | "reserved";

export function nameError(
  flows: AuthenticationFlowsConfig | undefined,
  types: readonly FlowType[],
  name: string,
  allowDefault: boolean
): NameError | null {
  if (!NAME_PATTERN.test(name)) {
    return "invalid";
  }
  if (
    types.some((t) =>
      (flows?.[flowListKey(t)] ?? []).some((f) => f.name === name)
    )
  ) {
    return "taken";
  }
  if (name === DEFAULT_FLOW_NAME && !allowDefault) {
    return "reserved";
  }
  return null;
}

// The first of base_2, base_3, … that no flow of the types has.
export function freeName(
  flows: AuthenticationFlowsConfig | undefined,
  types: readonly FlowType[],
  base: string
): string {
  for (let n = 2; ; n++) {
    const name = `${base}_${n}`;
    if (nameError(flows, types, name, false) == null) {
      return name;
    }
  }
}
