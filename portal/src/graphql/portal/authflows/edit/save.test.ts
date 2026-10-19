import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import {
  flowPointer,
  rawFlow,
  replaceFlow,
  removeFlows,
  replaceFlows,
  stepErrors,
} from "./save";

const raw: PortalAPIAppConfig = {
  id: "app",
  x_other: { kept: true },
  authentication_flow: {
    login_flows: [
      {
        name: "default",
        steps: [{ type: "identify", one_of: [], x_unknown: 1 } as never],
      },
      { name: "vip", steps: [{ type: "check_account_status" }] },
    ],
    signup_flows: [{ name: "default", steps: [{ type: "identify" }] }],
  },
} as PortalAPIAppConfig;

describe("rawFlow", () => {
  it("finds a flow by type and name", () => {
    expect(rawFlow(raw, "login", "vip")?.steps?.[0].type).toEqual(
      "check_account_status"
    );
    expect(rawFlow(raw, "reauth", "default")).toBeNull();
  });
});

describe("replaceFlow", () => {
  it("writes back an unchanged flow as an equal config", () => {
    const flow = rawFlow(raw, "login", "default")!;
    expect(replaceFlow(raw, "login", "default", flow)).toEqual(raw);
  });

  it("replaces only the matching flow", () => {
    const next = replaceFlow(raw, "login", "vip", { name: "vip", steps: [] });
    expect(next.authentication_flow?.login_flows?.[1]).toEqual({
      name: "vip",
      steps: [],
    });
    expect(next.authentication_flow?.login_flows?.[0]).toBe(
      raw.authentication_flow?.login_flows?.[0]
    );
    expect(next.authentication_flow?.signup_flows).toBe(
      raw.authentication_flow?.signup_flows
    );
    expect((next as unknown as Record<string, unknown>).x_other).toEqual({
      kept: true,
    });
  });

  it("appends a flow that is not there yet, creating the lists", () => {
    const flow = { name: "default", steps: [{ type: "identify" }] };
    const next = replaceFlow({ id: "app" }, "reauth", "default", flow);
    expect(next.authentication_flow?.reauth_flows).toEqual([flow]);
  });
});

describe("flowPointer", () => {
  it("points at the flow, or where it will be appended", () => {
    expect(flowPointer(raw, "login", "vip")).toEqual(
      "/authentication_flow/login_flows/1"
    );
    expect(flowPointer(raw, "login", "new")).toEqual(
      "/authentication_flow/login_flows/2"
    );
    expect(flowPointer(raw, "reauth", "default")).toEqual(
      "/authentication_flow/reauth_flows/0"
    );
  });
});

describe("stepErrors", () => {
  const error = {
    errorName: "Invalid",
    reason: "ValidationFailed",
    info: {
      causes: [
        {
          location:
            "/authentication_flow/login_flows/1/steps/0/one_of/0/target_step",
          kind: "general",
          details: { msg: "target_step does not refer to any named step" },
        },
        {
          location: "/authentication_flow/login_flows/0/steps/0",
          kind: "general",
          details: { msg: "another flow" },
        },
        {
          location: "/http",
          kind: "general",
          details: { msg: "elsewhere" },
        },
      ],
    },
  };

  it("maps causes inside the flow to the deepest node", () => {
    const errors = stepErrors(error, "/authentication_flow/login_flows/1", [
      "/steps/0",
      "/steps/0/one_of/0/steps/0",
    ]);
    expect([...errors.keys()]).toEqual(["/steps/0"]);
    expect(errors.get("/steps/0")?.map((c) => c.location)).toEqual([
      "/authentication_flow/login_flows/1/steps/0/one_of/0/target_step",
    ]);
  });

  it("does not match a sibling index by prefix", () => {
    const errors = stepErrors(error, "/authentication_flow/login_flows/1", [
      "/steps/1",
    ]);
    expect(errors.size).toEqual(0);
  });

  it("ignores errors that are not validation errors", () => {
    expect(
      stepErrors(new Error("network"), "/authentication_flow/login_flows/1", [
        "/steps/0",
      ]).size
    ).toEqual(0);
  });
});

describe("replaceFlows", () => {
  it("replaces or appends every draft", () => {
    const base = {
      authentication_flow: { login_flows: [{ name: "a", steps: [] }] },
    } as unknown as PortalAPIAppConfig;
    const next = replaceFlows(base, [
      {
        type: "login",
        name: "a",
        flow: { name: "a", steps: [{ type: "identify" }] },
      },
      { type: "signup", name: "a", flow: { name: "a", steps: [] } },
    ]);
    expect(next.authentication_flow?.login_flows?.[0].steps).toHaveLength(1);
    expect(next.authentication_flow?.signup_flows?.map((f) => f.name)).toEqual([
      "a",
    ]);
  });
});

describe("removeFlows", () => {
  it("removes the named flows and keeps the rest", () => {
    const next = removeFlows(raw, [{ type: "login", name: "default" }]);
    expect(next.authentication_flow?.login_flows?.map((f) => f.name)).toEqual([
      "vip",
    ]);
    expect(next.authentication_flow?.signup_flows).toHaveLength(1);
  });
});

describe("removeFlows on the last flow of a type", () => {
  it("drops the type's list rather than leaving it empty", () => {
    const next = removeFlows(raw, [{ type: "signup", name: "default" }]);
    expect(next.authentication_flow).not.toHaveProperty("signup_flows");
  });
});
