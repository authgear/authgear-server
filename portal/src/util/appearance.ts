// Portal appearance: light, dark, or follow the device.
//
// The resolved mode is applied as a `light-theme` or `dark-theme` class on
// <html>, which Radix Themes, Tailwind and content portalled to <body> all pick
// up. Radix accepts `dark`/`light` too, but `components/v2` already uses a bare
// `dark` class to mark a single component as sitting on a dark island (see the
// `darkMode` props there), so the page-level class has to be the other spelling
// or those two meanings collide. The inline bootstrap script in index.html
// applies the same class before the bundle loads so the first paint is already
// correct; keep APPEARANCE_STORAGE_KEY and the resolve rule in sync with it.

export const APPEARANCE_STORAGE_KEY = "authgear-portal-appearance";
const DARK_MEDIA_QUERY = "(prefers-color-scheme: dark)";

export type Appearance = "light" | "dark" | "system";
export type ResolvedAppearance = "light" | "dark";

export interface AppearanceState {
  preference: Appearance;
  resolved: ResolvedAppearance;
}

export function parseAppearance(raw: string | null | undefined): Appearance {
  switch (raw) {
    case "light":
    case "dark":
    case "system":
      return raw;
    default:
      return "system";
  }
}

export function resolveAppearance(
  preference: Appearance,
  systemPrefersDark: boolean
): ResolvedAppearance {
  switch (preference) {
    case "light":
      return "light";
    case "dark":
      return "dark";
    case "system":
      return systemPrefersDark ? "dark" : "light";
  }
}

export function applyResolvedAppearance(
  root: Element,
  resolved: ResolvedAppearance
): void {
  root.classList.remove("light-theme", "dark-theme");
  root.classList.add(`${resolved}-theme`);
}

function readStoredAppearance(): Appearance {
  try {
    return parseAppearance(window.localStorage.getItem(APPEARANCE_STORAGE_KEY));
  } catch {
    // Storage can be unavailable (privacy mode, disabled site data).
    return "system";
  }
}

function systemPrefersDark(): boolean {
  return window.matchMedia(DARK_MEDIA_QUERY).matches;
}

let state: AppearanceState | null = null;
const listeners = new Set<() => void>();

export function getAppearanceState(): AppearanceState {
  if (state == null) {
    const preference = readStoredAppearance();
    state = {
      preference,
      resolved: resolveAppearance(preference, systemPrefersDark()),
    };
  }
  return state;
}

function commit(preference: Appearance): void {
  const prev = getAppearanceState();
  const next: AppearanceState = {
    preference,
    resolved: resolveAppearance(preference, systemPrefersDark()),
  };
  const unchanged =
    prev.preference === next.preference && prev.resolved === next.resolved;
  state = next;
  // Applied even when unchanged, so the class is correct however the store
  // was first reached.
  applyResolvedAppearance(document.documentElement, next.resolved);
  if (unchanged) {
    return;
  }
  for (const listener of listeners) {
    listener();
  }
}

export function setAppearance(preference: Appearance): void {
  // Pin the current state before writing: commit() initializes lazily from
  // storage, so writing first would make it read back the new value as the
  // old one and skip notifying subscribers.
  getAppearanceState();
  try {
    window.localStorage.setItem(APPEARANCE_STORAGE_KEY, preference);
  } catch {
    // Still apply for this page even if the choice cannot be persisted.
  }
  commit(preference);
}

export function subscribeAppearance(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function onExternalChange(): void {
  commit(readStoredAppearance());
}

let initialized = false;

// Applies the current mode and keeps it in sync with the OS setting and with
// changes made in other tabs. Call once before rendering the app.
export function initAppearance(): void {
  if (initialized) {
    return;
  }
  initialized = true;
  applyResolvedAppearance(
    document.documentElement,
    getAppearanceState().resolved
  );
  window
    .matchMedia(DARK_MEDIA_QUERY)
    .addEventListener("change", onExternalChange);
  window.addEventListener("storage", (e: StorageEvent) => {
    // key is null when the whole storage was cleared.
    if (e.key === null || e.key === APPEARANCE_STORAGE_KEY) {
      onExternalChange();
    }
  });
}
