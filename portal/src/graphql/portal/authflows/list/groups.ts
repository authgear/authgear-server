import { produce } from "immer";
import {
  AuthenticationFlowReference,
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
  UIAuthenticationFlowGroup,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey, flowTypes } from "../model";
import { NAME_PATTERN, NameError } from "./actions";

// Flow groups (ui.authentication_flow.groups): Auth UI runs the group's flow
// of each type, and applications allow groups by name
// (oauth.clients[].x_authentication_flow_allowlist.groups).

// Undefined, the default group runs the default flows; it is never a name
// picked for a new or renamed group.
export const DEFAULT_GROUP_NAME = "default";

// The flow a group runs for each type.
export type GroupSelection = Record<FlowType, string>;

export type GroupAction =
  | { kind: "add"; name: string; selection: GroupSelection }
  | { kind: "set-flows"; name: string; selection: GroupSelection }
  | { kind: "rename"; name: string; newName: string }
  | { kind: "delete"; name: string };

export function flowGroups(
  config: PortalAPIAppConfig | null | undefined
): UIAuthenticationFlowGroup[] {
  return config?.ui?.authentication_flow?.groups ?? [];
}

// The flows a group can pick for a type: default, which always exists, then
// the custom ones.
export function flowNamesOfType(
  flows: AuthenticationFlowsConfig | undefined,
  type: FlowType
): string[] {
  const names = (flows?.[flowListKey(type)] ?? [])
    .map((f) => f.name)
    .filter((n) => n !== DEFAULT_FLOW_NAME);
  return [DEFAULT_FLOW_NAME, ...names];
}

// Auth UI runs a group's first flow of a type. A group without one fails
// that page, so a missing type reads as default and is saved as default.
export function groupSelection(
  group: UIAuthenticationFlowGroup | null
): GroupSelection {
  const out = {} as GroupSelection;
  for (const type of flowTypes) {
    out[type] =
      group?.flows?.find((f) => f.type === type)?.name ?? DEFAULT_FLOW_NAME;
  }
  return out;
}

// The applications that allow the group, by name.
export function groupUsers(config: PortalAPIAppConfig, name: string): string[] {
  return (config.oauth?.clients ?? [])
    .filter((c) =>
      (c.x_authentication_flow_allowlist?.groups ?? []).some(
        (g) => g.name === name
      )
    )
    .map((c) => c.name ?? c.client_id);
}

// The applications that run the default group: those allowing no group, which
// get it when they ask for none, and those that list it. Machine-to-machine
// applications sign no one in.
export function standardGroupUsers(config: PortalAPIAppConfig): string[] {
  return (config.oauth?.clients ?? [])
    .filter((c) => c.x_application_type !== "m2m")
    .filter((c) => {
      const groups = c.x_authentication_flow_allowlist?.groups ?? [];
      return (
        groups.length === 0 || groups.some((g) => g.name === DEFAULT_GROUP_NAME)
      );
    })
    .map((c) => c.name ?? c.client_id);
}

// Applications whose allowlist is this group alone. Deleting the group would
// empty it, and an empty allowlist allows every flow.
export function groupDeleteBlockers(
  config: PortalAPIAppConfig,
  name: string
): string[] {
  return (config.oauth?.clients ?? [])
    .filter((c) => {
      const groups = c.x_authentication_flow_allowlist?.groups ?? [];
      const flows = c.x_authentication_flow_allowlist?.flows ?? [];
      return (
        flows.length === 0 &&
        groups.length > 0 &&
        groups.every((g) => g.name === name)
      );
    })
    .map((c) => c.name ?? c.client_id);
}

// current: the group being renamed, whose own name is not taken.
export function groupNameError(
  config: PortalAPIAppConfig | null | undefined,
  name: string,
  current: string | null
): NameError | null {
  if (!NAME_PATTERN.test(name)) {
    return "invalid";
  }
  if (name === DEFAULT_GROUP_NAME) {
    return "reserved";
  }
  if (name !== current && flowGroups(config).some((g) => g.name === name)) {
    return "taken";
  }
  return null;
}

// refs with the first flow of each type set to the selection. Further flows
// of a type are kept: Auth UI ignores them, the Authflow API allows them.
function withSelection(
  refs: AuthenticationFlowReference[],
  selection: GroupSelection
): AuthenticationFlowReference[] {
  const out = refs.map((r) => ({ ...r }));
  for (const type of flowTypes) {
    const first = out.find((r) => r.type === type);
    if (first != null) {
      first.name = selection[type];
    } else {
      out.push({ type, name: selection[type] });
    }
  }
  return out;
}

export function applyGroupAction(
  config: PortalAPIAppConfig,
  action: GroupAction
): PortalAPIAppConfig {
  return produce(config, (d) => {
    // Only adding writes the section; the other actions need a group.
    if (action.kind === "add") {
      d.ui ??= {};
      d.ui.authentication_flow ??= {};
      (d.ui.authentication_flow.groups ??= []).push({
        name: action.name,
        flows: withSelection([], action.selection),
      });
      return;
    }
    const section = d.ui?.authentication_flow;
    const groups = section?.groups ?? [];
    const allowlists = (d.oauth?.clients ?? []).flatMap((c) =>
      c.x_authentication_flow_allowlist != null
        ? [c.x_authentication_flow_allowlist]
        : []
    );
    switch (action.kind) {
      case "set-flows": {
        const group = groups.find((g) => g.name === action.name);
        if (group != null) {
          group.flows = withSelection(group.flows ?? [], action.selection);
        }
        break;
      }
      case "rename":
        for (const g of groups) {
          if (g.name === action.name) {
            g.name = action.newName;
          }
        }
        for (const a of allowlists) {
          for (const g of a.groups ?? []) {
            if (g.name === action.name) {
              g.name = action.newName;
            }
          }
        }
        break;
      case "delete":
        if (section != null) {
          section.groups = groups.filter((g) => g.name !== action.name);
        }
        for (const a of allowlists) {
          if (a.groups != null) {
            a.groups = a.groups.filter((g) => g.name !== action.name);
          }
        }
        break;
    }
  });
}
