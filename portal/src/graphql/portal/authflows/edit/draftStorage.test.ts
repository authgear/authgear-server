import { afterEach, describe, expect, it, jest } from "@jest/globals";
import { readStoredDraft, removeStoredDraft, storeDraft } from "./draftStorage";

const flow = { name: "vip", steps: [{ type: "identify" as const }] };

afterEach(() => {
  window.localStorage.clear();
  jest.restoreAllMocks();
});

describe("storeDraft", () => {
  it("keeps a draft per project and flow", () => {
    storeDraft("app", "login", "vip", flow, 1000);
    expect(readStoredDraft("app", "login", "vip")).toEqual({
      flow,
      savedAt: 1000,
    });
    expect(readStoredDraft("app", "signup", "vip")).toBeNull();
    expect(readStoredDraft("other", "login", "vip")).toBeNull();
  });

  it("is gone once removed", () => {
    storeDraft("app", "login", "vip", flow, 1000);
    removeStoredDraft("app", "login", "vip");
    expect(readStoredDraft("app", "login", "vip")).toBeNull();
  });
});

describe("readStoredDraft", () => {
  it("ignores what it cannot read as a draft", () => {
    window.localStorage.setItem(
      "authgear-portal-authflow-draft:app:login:vip",
      "{not json"
    );
    expect(readStoredDraft("app", "login", "vip")).toBeNull();
    window.localStorage.setItem(
      "authgear-portal-authflow-draft:app:login:vip",
      JSON.stringify({ savedAt: 1 })
    );
    expect(readStoredDraft("app", "login", "vip")).toBeNull();
  });

  it("does not throw when storage is unavailable", () => {
    jest.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("denied");
    });
    jest.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("denied");
    });
    expect(() => storeDraft("app", "login", "vip", flow, 1)).not.toThrow();
    expect(readStoredDraft("app", "login", "vip")).toBeNull();
  });
});
