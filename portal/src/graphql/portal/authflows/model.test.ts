import { describe, it, expect } from "@jest/globals";
import {
  DEFAULT_FLOW_NAME,
  DEFAULT_FLOW_TYPE,
  flowSelectionSearch,
  listFlowNames,
  parseFlowSelection,
  resolveFlow,
} from "./model";
import {
  customSignupWithForm,
  generatedEmailOAuthPasskey2FA,
} from "./__fixtures__/flows";
import { AuthenticationFlowsConfig } from "../../../types";

const customWithNamed: AuthenticationFlowsConfig = {
  signup_flows: [
    ...(customSignupWithForm.signup_flows ?? []),
    { name: "vip", steps: [] },
  ],
};

describe("resolveFlow", () => {
  it("prefers the custom flow with the name", () => {
    const resolved = resolveFlow(
      customSignupWithForm,
      generatedEmailOAuthPasskey2FA,
      "signup",
      DEFAULT_FLOW_NAME
    );
    expect(resolved?.source).toEqual("custom");
    expect(resolved?.flow).toBe(customSignupWithForm.signup_flows?.[0]);
  });

  it("falls back to the generated flow for default", () => {
    const resolved = resolveFlow(
      customSignupWithForm,
      generatedEmailOAuthPasskey2FA,
      "login",
      DEFAULT_FLOW_NAME
    );
    expect(resolved?.source).toEqual("generated");
    expect(resolved?.flow).toBe(generatedEmailOAuthPasskey2FA.login_flows?.[0]);
  });

  it("returns null for a missing non-default name", () => {
    expect(
      resolveFlow(
        customSignupWithForm,
        generatedEmailOAuthPasskey2FA,
        "login",
        "missing"
      )
    ).toBeNull();
  });
});

describe("listFlowNames", () => {
  it("lists default once, then custom names in config order", () => {
    expect(listFlowNames(customWithNamed, "signup")).toEqual([
      "default",
      "vip",
    ]);
  });

  it("lists only default when there is no custom flow", () => {
    expect(listFlowNames(undefined, "login")).toEqual(["default"]);
  });
});

describe("parseFlowSelection", () => {
  it("falls back for an unknown type", () => {
    expect(
      parseFlowSelection(new URLSearchParams("type=bogus"), customWithNamed)
    ).toEqual({ type: DEFAULT_FLOW_TYPE, name: DEFAULT_FLOW_NAME });
  });

  it("keeps a custom name of the selected type", () => {
    expect(
      parseFlowSelection(
        new URLSearchParams("type=signup&name=vip"),
        customWithNamed
      )
    ).toEqual({ type: "signup", name: "vip" });
  });

  it("falls back to default for a name that exists only for another type", () => {
    expect(
      parseFlowSelection(
        new URLSearchParams("type=login&name=vip"),
        customWithNamed
      )
    ).toEqual({ type: "login", name: DEFAULT_FLOW_NAME });
  });
});

describe("flowSelectionSearch", () => {
  it("serializes type and name", () => {
    expect(flowSelectionSearch({ type: "login", name: "default" })).toEqual(
      "?type=login&name=default"
    );
  });
});
