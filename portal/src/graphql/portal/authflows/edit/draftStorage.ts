import { AuthenticationFlowConfig } from "../../../../types";
import { FlowType } from "../model";

// An unsaved flow draft kept in this browser, so a refresh or a closed tab
// does not lose it. Storage can be unavailable (privacy mode, disabled site
// data, quota), and then drafts are simply not kept.

export interface StoredDraft {
  flow: AuthenticationFlowConfig;
  // When it was stored, in milliseconds since the epoch.
  savedAt: number;
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
  now: number
): void {
  const value: StoredDraft = { flow, savedAt: now };
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
    const { flow, savedAt } = parsed as Record<string, unknown>;
    if (
      typeof savedAt !== "number" ||
      typeof flow !== "object" ||
      flow == null
    ) {
      return null;
    }
    return { flow: flow as AuthenticationFlowConfig, savedAt };
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
