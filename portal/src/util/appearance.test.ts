import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  jest,
} from "@jest/globals";
import * as fs from "fs";
import * as path from "path";
import {
  APPEARANCE_STORAGE_KEY,
  applyResolvedAppearance,
  parseAppearance,
  resolveAppearance,
} from "./appearance";

type AppearanceModule = typeof import("./appearance");

interface MatchMediaMock {
  matches: boolean;
  listeners: Set<() => void>;
}

function installMatchMedia(prefersDark: boolean): MatchMediaMock {
  const mock: MatchMediaMock = { matches: prefersDark, listeners: new Set() };
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    writable: true,
    value: (query: string) => ({
      media: query,
      get matches() {
        return mock.matches;
      },
      addEventListener: (_type: string, listener: () => void) => {
        mock.listeners.add(listener);
      },
      removeEventListener: (_type: string, listener: () => void) => {
        mock.listeners.delete(listener);
      },
    }),
  });
  return mock;
}

function loadFreshModule(): AppearanceModule {
  let mod: AppearanceModule | undefined;
  jest.isolateModules(() => {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    mod = require("./appearance") as AppearanceModule;
  });
  if (mod === undefined) {
    throw new Error("failed to load module");
  }
  return mod;
}

describe("parseAppearance", () => {
  it("accepts the three known values", () => {
    expect(parseAppearance("light")).toBe("light");
    expect(parseAppearance("dark")).toBe("dark");
    expect(parseAppearance("system")).toBe("system");
  });

  it("falls back to system for missing or unknown values", () => {
    expect(parseAppearance(null)).toBe("system");
    expect(parseAppearance(undefined)).toBe("system");
    expect(parseAppearance("")).toBe("system");
    expect(parseAppearance("Dark")).toBe("system");
    expect(parseAppearance("auto")).toBe("system");
  });
});

describe("resolveAppearance", () => {
  it("returns explicit preferences unchanged", () => {
    expect(resolveAppearance("light", true)).toBe("light");
    expect(resolveAppearance("dark", false)).toBe("dark");
  });

  it("follows the device for system", () => {
    expect(resolveAppearance("system", true)).toBe("dark");
    expect(resolveAppearance("system", false)).toBe("light");
  });
});

describe("applyResolvedAppearance", () => {
  it("keeps exactly one appearance class and leaves others alone", () => {
    const el = document.createElement("div");
    el.classList.add("keep-me", "light-theme");
    applyResolvedAppearance(el, "dark");
    expect(Array.from(el.classList)).toEqual(["keep-me", "dark-theme"]);
    applyResolvedAppearance(el, "light");
    expect(Array.from(el.classList)).toEqual(["keep-me", "light-theme"]);
  });
});

describe("appearance store", () => {
  let media: MatchMediaMock;

  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.className = "";
    media = installMatchMedia(false);
  });

  afterEach(() => {
    window.localStorage.clear();
    document.documentElement.className = "";
  });

  it("defaults to system and follows the device", () => {
    media.matches = true;
    const mod = loadFreshModule();
    expect(mod.getAppearanceState()).toEqual({
      preference: "system",
      resolved: "dark",
    });
  });

  it("reads a stored preference", () => {
    window.localStorage.setItem(APPEARANCE_STORAGE_KEY, "dark");
    const mod = loadFreshModule();
    expect(mod.getAppearanceState()).toEqual({
      preference: "dark",
      resolved: "dark",
    });
  });

  it("persists, applies and notifies on setAppearance", () => {
    const mod = loadFreshModule();
    mod.initAppearance();
    expect(document.documentElement.classList.contains("light-theme")).toBe(
      true
    );

    const listener = jest.fn<() => void>();
    mod.subscribeAppearance(listener);

    mod.setAppearance("dark");
    expect(window.localStorage.getItem(APPEARANCE_STORAGE_KEY)).toBe("dark");
    expect(mod.getAppearanceState()).toEqual({
      preference: "dark",
      resolved: "dark",
    });
    expect(document.documentElement.classList.contains("dark-theme")).toBe(
      true
    );
    expect(document.documentElement.classList.contains("light-theme")).toBe(
      false
    );
    expect(listener).toHaveBeenCalledTimes(1);

    // Setting the same value again is a no-op for subscribers.
    mod.setAppearance("dark");
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("re-resolves system when the device setting changes", () => {
    const mod = loadFreshModule();
    mod.initAppearance();
    const listener = jest.fn<() => void>();
    mod.subscribeAppearance(listener);

    media.matches = true;
    for (const l of media.listeners) {
      l();
    }
    expect(mod.getAppearanceState().resolved).toBe("dark");
    expect(document.documentElement.classList.contains("dark-theme")).toBe(
      true
    );
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("ignores device changes while an explicit preference is set", () => {
    window.localStorage.setItem(APPEARANCE_STORAGE_KEY, "light");
    const mod = loadFreshModule();
    mod.initAppearance();
    const listener = jest.fn<() => void>();
    mod.subscribeAppearance(listener);

    media.matches = true;
    for (const l of media.listeners) {
      l();
    }
    expect(mod.getAppearanceState().resolved).toBe("light");
    expect(listener).not.toHaveBeenCalled();
  });

  it("applies the class when setAppearance is the first call", () => {
    // Storybook reaches the store through setAppearance() without calling
    // initAppearance() first; the class must still land on <html>.
    const mod = loadFreshModule();
    mod.setAppearance("dark");
    expect(document.documentElement.classList.contains("dark-theme")).toBe(
      true
    );
  });

  it("picks up changes made in another tab", () => {
    const mod = loadFreshModule();
    mod.initAppearance();

    window.localStorage.setItem(APPEARANCE_STORAGE_KEY, "dark");
    window.dispatchEvent(
      new StorageEvent("storage", { key: APPEARANCE_STORAGE_KEY })
    );
    expect(mod.getAppearanceState().preference).toBe("dark");
    expect(document.documentElement.classList.contains("dark-theme")).toBe(
      true
    );

    window.localStorage.clear();
    window.dispatchEvent(new StorageEvent("storage", { key: null }));
    expect(mod.getAppearanceState().preference).toBe("system");
  });
});

describe("appearance classes", () => {
  it("does not use the bare class that marks a dark island", () => {
    // `components/v2` puts a plain `dark` class on a single component to render
    // it on a dark surface; the page-level class must stay distinct from it.
    const el = document.createElement("div");
    applyResolvedAppearance(el, "dark");
    expect(el.classList.contains("dark")).toBe(false);
    expect(el.classList.contains("dark-theme")).toBe(true);
  });
});

describe("index.html bootstrap script", () => {
  it("uses the same storage key as the module", () => {
    const html = fs.readFileSync(
      path.join(__dirname, "..", "index.html"),
      "utf8"
    );
    expect(html).toContain(`"${APPEARANCE_STORAGE_KEY}"`);
  });
});
