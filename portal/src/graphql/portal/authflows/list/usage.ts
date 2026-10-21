import { produce } from "immer";
import { PortalAPIAppConfig } from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType } from "../model";

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
      x_authentication_flow_allowlist?: { flows?: FlowRef[] };
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
