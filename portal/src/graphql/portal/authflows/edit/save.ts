import { produce } from "immer";
import {
  AuthenticationFlowConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { isAPIError } from "../../../../error/error";
import { parseRawError } from "../../../../error/parse";
import { ValidationFailedErrorInfoCause } from "../../../../error/validation";
import { FlowType, flowListKey } from "../model";
import { isWithin } from "./pointer";
import { applyFlowAction } from "../list/actions";
import { renameFlowUses } from "../list/usage";

// One flow being edited.
export interface EditorDraft {
  type: FlowType;
  name: string;
  flow: AuthenticationFlowConfig;
}

// The flows being edited: one, or a bundle's three. Empty while viewing.
export interface EditorState {
  drafts: EditorDraft[];
  // Flows to remove on save: an edited default reset to the generated one.
  removals: { type: FlowType; name: string }[];
  // A flow to rename on save.
  rename?: { type: FlowType; name: string; newName: string };
}

function flowList(
  raw: PortalAPIAppConfig,
  type: FlowType
): AuthenticationFlowConfig[] {
  return raw.authentication_flow?.[flowListKey(type)] ?? [];
}

// The flow as written in authgear.yaml, before defaults are applied.
export function rawFlow(
  raw: PortalAPIAppConfig,
  type: FlowType,
  name: string
): AuthenticationFlowConfig | null {
  return flowList(raw, type).find((f) => f.name === name) ?? null;
}

// raw with the flow of this type and name replaced, or appended when there
// is none. Nothing else changes.
export function replaceFlow(
  raw: PortalAPIAppConfig,
  type: FlowType,
  name: string,
  flow: AuthenticationFlowConfig
): PortalAPIAppConfig {
  return produce(raw, (d) => {
    d.authentication_flow ??= {};
    const key = flowListKey(type);
    const list = (d.authentication_flow[key] ??= []);
    const i = list.findIndex((f) => f.name === name);
    if (i >= 0) {
      list[i] = flow;
    } else {
      list.push(flow);
    }
  });
}

// raw with every draft written in, as replaceFlow does for one.
export function replaceFlows(
  raw: PortalAPIAppConfig,
  drafts: EditorDraft[]
): PortalAPIAppConfig {
  return drafts.reduce((r, d) => replaceFlow(r, d.type, d.name, d.flow), raw);
}

// raw without the named flows; removing a default flow resets it to the
// generated one.
export function removeFlows(
  raw: PortalAPIAppConfig,
  removals: { type: FlowType; name: string }[]
): PortalAPIAppConfig {
  return produce(raw, (d) => {
    for (const { type, name } of removals) {
      const key = flowListKey(type);
      const list = d.authentication_flow?.[key];
      if (d.authentication_flow != null && list != null) {
        const rest = list.filter((f) => f.name !== name);
        // A flow list must have at least one flow, or be absent.
        if (rest.length > 0) {
          d.authentication_flow[key] = rest;
        } else {
          delete d.authentication_flow[key];
        }
      }
    }
  });
}

// raw with the flow renamed, along with the signup_login hand-offs, flow
// groups and applications that name it.
export function renameFlow(
  raw: PortalAPIAppConfig,
  type: FlowType,
  name: string,
  newName: string
): PortalAPIAppConfig {
  const next = produce(raw, (d) => {
    d.authentication_flow = applyFlowAction(raw.authentication_flow, {
      kind: "rename",
      type,
      name,
      newName,
    });
  });
  return renameFlowUses(next, type, name, newName);
}

// Where the flow is, or will be once appended, in the config: the prefix of
// the save errors that belong to it.
export function flowPointer(
  raw: PortalAPIAppConfig,
  type: FlowType,
  name: string
): string {
  const list = flowList(raw, type);
  const i = list.findIndex((f) => f.name === name);
  return `/authentication_flow/${flowListKey(type)}/${
    i >= 0 ? i : list.length
  }`;
}

// The validation causes of a save error that fall inside the flow, keyed by
// the node whose pointer is the longest prefix of their location; causes on
// the flow itself, outside any step, are under "".
export function stepErrors(
  error: unknown,
  prefix: string,
  nodeIDs: string[]
): Map<string, ValidationFailedErrorInfoCause[]> {
  const out = new Map<string, ValidationFailedErrorInfoCause[]>();
  const apiErrors = isAPIError(error) ? [error] : parseRawError(error);
  for (const e of apiErrors) {
    if (e.reason !== "ValidationFailed") {
      continue;
    }
    for (const cause of e.info.causes) {
      const location = cause.location;
      if (!isWithin(location, prefix)) {
        continue;
      }
      const inFlow = location.slice(prefix.length);
      let best: string | null = null;
      for (const id of nodeIDs) {
        if (isWithin(inFlow, id) && (best == null || id.length > best.length)) {
          best = id;
        }
      }
      const key = best ?? "";
      out.set(key, [...(out.get(key) ?? []), cause]);
    }
  }
  return out;
}
