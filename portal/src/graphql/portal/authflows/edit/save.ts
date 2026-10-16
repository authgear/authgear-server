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

// The draft being edited; null while only viewing.
export interface EditorDraft {
  type: FlowType;
  name: string;
  flow: AuthenticationFlowConfig;
}

export interface EditorState {
  draft: EditorDraft | null;
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
// the node whose pointer is the longest prefix of their location.
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
      if (best != null) {
        out.set(best, [...(out.get(best) ?? []), cause]);
      }
    }
  }
  return out;
}
