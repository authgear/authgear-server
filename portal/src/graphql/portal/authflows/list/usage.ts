import { produce } from "immer";
import { PortalAPIAppConfig } from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType } from "../model";
import { deleteBlockers } from "./actions";

// Other settings that name flows: flow groups (ui.authentication_flow.groups)
// and applications' allowed flows (oauth.clients[].
// x_authentication_flow_allowlist.flows). The server refuses a config whose
// references name a flow that does not exist, except default, which the
// generated flow always answers to. The portal's types leave these out.

interface FlowRef {
  type: string;
  name: string;
}

interface WithFlowRefs {
  ui?: {
    authentication_flow?: {
      groups?: { name: string; flows?: FlowRef[] }[];
    };
  };
  oauth?: {
    clients?: {
      client_id: string;
      name?: string;
      x_authentication_flow_allowlist?: {
        groups?: { name: string }[];
        flows?: FlowRef[];
      };
    }[];
  };
}

export type FlowUse =
  | { kind: "group"; name: string }
  | { kind: "application"; name: string };

function names(ref: FlowRef, type: FlowType, name: string): boolean {
  return ref.type === type && ref.name === name;
}

// The groups and applications that name the flow.
export function flowUses(
  config: PortalAPIAppConfig,
  type: FlowType,
  name: string
): FlowUse[] {
  if (name === DEFAULT_FLOW_NAME) {
    return [];
  }
  const c = config as WithFlowRefs;
  const uses: FlowUse[] = [];
  for (const g of c.ui?.authentication_flow?.groups ?? []) {
    if ((g.flows ?? []).some((f) => names(f, type, name))) {
      uses.push({ kind: "group", name: g.name });
    }
  }
  for (const client of c.oauth?.clients ?? []) {
    const flows = client.x_authentication_flow_allowlist?.flows ?? [];
    if (flows.some((f) => names(f, type, name))) {
      uses.push({
        kind: "application",
        name: client.name ?? client.client_id,
      });
    }
  }
  return uses;
}

// Applications whose allowlist is this flow alone. Deleting the flow would
// empty it, and an empty allowlist allows every flow.
export function onlyAllowedBy(
  config: PortalAPIAppConfig,
  type: FlowType,
  name: string
): string[] {
  const c = config as WithFlowRefs;
  return (c.oauth?.clients ?? [])
    .filter((client) => {
      const allowlist = client.x_authentication_flow_allowlist;
      const flows = allowlist?.flows ?? [];
      return (
        (allowlist?.groups ?? []).length === 0 &&
        flows.length > 0 &&
        flows.every((f) => names(f, type, name))
      );
    })
    .map((client) => client.name ?? client.client_id);
}

// config with every group and application reference to the flow renamed.
export function renameFlowUses(
  config: PortalAPIAppConfig,
  type: FlowType,
  from: string,
  to: string
): PortalAPIAppConfig {
  return produce(config, (d) => {
    const c = d as WithFlowRefs;
    const refs = [
      ...(c.ui?.authentication_flow?.groups ?? []).flatMap(
        (g) => g.flows ?? []
      ),
      ...(c.oauth?.clients ?? []).flatMap(
        (client) => client.x_authentication_flow_allowlist?.flows ?? []
      ),
    ];
    for (const ref of refs) {
      if (names(ref, type, from)) {
        ref.name = to;
      }
    }
  });
}

// config with the flow taken out of groups and applications, for deleting
// it: a group falls back to the default flow of that type, which always
// exists, and an application no longer lists the flow as allowed.
export function removeFlowUses(
  config: PortalAPIAppConfig,
  type: FlowType,
  name: string
): PortalAPIAppConfig {
  return produce(config, (d) => {
    const c = d as WithFlowRefs;
    for (const g of c.ui?.authentication_flow?.groups ?? []) {
      for (const ref of g.flows ?? []) {
        if (names(ref, type, name)) {
          ref.name = DEFAULT_FLOW_NAME;
        }
      }
    }
    for (const client of c.oauth?.clients ?? []) {
      const allowlist = client.x_authentication_flow_allowlist;
      if (allowlist?.flows != null) {
        allowlist.flows = allowlist.flows.filter(
          (ref) => !names(ref, type, name)
        );
      }
    }
  });
}

// The settings that name the flow, which a rename updates with it; null
// when none do.
export function renameNotice(
  renderToString: (id: string, values?: Record<string, string>) => string,
  config: PortalAPIAppConfig | null,
  type: FlowType,
  name: string
): string | null {
  if (config == null) {
    return null;
  }
  const uses = flowUses(config, type, name);
  const groups = uses.filter((u) => u.kind === "group").map((u) => u.name);
  const apps = uses.filter((u) => u.kind === "application").map((u) => u.name);
  const handoffs = deleteBlockers(config.authentication_flow, type, name);
  const lines = [
    groups.length > 0
      ? renderToString("AuthFlowsScreen.list.rename.effect.groups", {
          names: groups.join(", "),
        })
      : null,
    apps.length > 0
      ? renderToString("AuthFlowsScreen.list.rename.effect.applications", {
          names: apps.join(", "),
        })
      : null,
    handoffs.length > 0
      ? renderToString("AuthFlowsScreen.list.rename.effect.handoffs", {
          names: handoffs.join(", "),
        })
      : null,
  ].filter((t): t is string => t != null);
  return lines.length > 0 ? lines.join(" ") : null;
}
