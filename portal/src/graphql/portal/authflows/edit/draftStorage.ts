import { AuthenticationFlowConfig } from "../../../../types";
import { FlowType } from "../model";
import { FormNameOp, TextEdits } from "./formText";

// An unsaved flow draft kept in this browser, so a refresh or a closed tab
// does not lose it. Storage can be unavailable (privacy mode, disabled site
// data, quota), and then drafts are simply not kept.

export interface StoredDraft {
  flow: AuthenticationFlowConfig;
  // When it was stored, in milliseconds since the epoch.
  savedAt: number;
  // Unsaved form text, kept with the flow.
  text?: StoredFormText;
}

export interface StoredFormText {
  edits: TextEdits;
  ops: FormNameOp[];
}

const KEY_PREFIX = "authgear-portal-authflow-draft";

function storageKey(appID: string, type: FlowType, name: string): string {
  return `${KEY_PREFIX}:${appID}:${type}:${name}`;
}

export function storeDraft(
  appID: string,
  type: FlowType,
  name: string,
  flow: AuthenticationFlowConfig,
  now: number,
  text?: StoredFormText
): void {
  const value: StoredDraft = { flow, savedAt: now, text };
  try {
    window.localStorage.setItem(
      storageKey(appID, type, name),
      JSON.stringify(value)
    );
  } catch {
    // Not kept; the leave-page warning still guards the draft.
  }
}

export function readStoredDraft(
  appID: string,
  type: FlowType,
  name: string
): StoredDraft | null {
  try {
    const raw = window.localStorage.getItem(storageKey(appID, type, name));
    if (raw == null) {
      return null;
    }
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== "object" || parsed == null) {
      return null;
    }
    const { flow, savedAt, text } = parsed as Record<string, unknown>;
    if (
      typeof savedAt !== "number" ||
      typeof flow !== "object" ||
      flow == null
    ) {
      return null;
    }
    return {
      flow: flow as AuthenticationFlowConfig,
      savedAt,
      text: isStoredFormText(text) ? text : undefined,
    };
  } catch {
    return null;
  }
}

export function removeStoredDraft(
  appID: string,
  type: FlowType,
  name: string
): void {
  try {
    window.localStorage.removeItem(storageKey(appID, type, name));
  } catch {
    // Nothing was kept, or nothing can be removed; either way it is unread.
  }
}

function isStoredFormText(value: unknown): value is StoredFormText {
  if (typeof value !== "object" || value == null) {
    return false;
  }
  const { edits, ops } = value as Record<string, unknown>;
  return typeof edits === "object" && edits != null && Array.isArray(ops);
}
