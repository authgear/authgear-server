import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowsConfig } from "../../../../types";
import {
  applyFlowAction,
  copyAs,
  deleteBlockers,
  freeName,
  nameError,
} from "./actions";

const handoff = (name: string, to: string) => ({
  name,
  steps: [
    {
      type: "identify",
      one_of: [{ identification: "email", login_flow: to, signup_flow: to }],
    },
  ],
});

const flows: AuthenticationFlowsConfig = {
  signup_login_flows: [
    handoff("default", "default"),
    handoff("members", "members"),
  ],
  signup_flows: [
    { name: "default", steps: [] },
    { name: "members", steps: [] },
  ],
  login_flows: [
    { name: "default", steps: [] },
    { name: "members", steps: [] },
  ],
};

const names = (
  f: AuthenticationFlowsConfig,
  key: keyof AuthenticationFlowsConfig
) => (f[key] ?? []).map((x) => x.name);
const targets = (f: AuthenticationFlowsConfig) =>
  (f.signup_login_flows ?? []).map((x) => [
    x.name,
    x.steps?.[0].one_of?.[0].login_flow,
  ]);

describe("set_default", () => {
  it("swaps a bundle's three flows with the defaults and their hand-offs", () => {
    const next = applyFlowAction(flows, {
      kind: "set_default",
      type: "signup_login",
      name: "members",
      bundle: true,
    });
    expect(names(next, "signup_flows")).toEqual(["members", "default"]);
    expect(targets(next)).toEqual([
      ["members", "members"],
      ["default", "default"],
    ]);
  });

  it("makes a flow default when only the generated one existed", () => {
    const next = applyFlowAction(
      { login_flows: [{ name: "staff", steps: [] }] },
      {
        kind: "set_default",
        type: "login",
        name: "staff",
        bundle: false,
      }
    );
    expect(names(next, "login_flows")).toEqual(["default"]);
  });

  it("rewrites hand-offs in other signup_login flows that named either flow", () => {
    const f: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("other", "default")],
      login_flows: [
        { name: "default", steps: [] },
        { name: "staff", steps: [] },
      ],
      signup_flows: [{ name: "default", steps: [] }],
    };
    const next = applyFlowAction(f, {
      kind: "set_default",
      type: "login",
      name: "staff",
      bundle: false,
    });
    expect(next.signup_login_flows?.[0].steps?.[0].one_of?.[0]).toMatchObject({
      login_flow: "staff",
      signup_flow: "default",
    });
  });
});

describe("rename", () => {
  it("renames a bundle and its hand-offs", () => {
    const next = applyFlowAction(flows, {
      kind: "rename",
      type: "signup_login",
      name: "members",
      newName: "staff",
      bundle: true,
    });
    expect(names(next, "login_flows")).toEqual(["default", "staff"]);
    expect(targets(next)[1]).toEqual(["staff", "staff"]);
  });
});

describe("add and copyAs", () => {
  it("copies a bundle under a new name with its hand-offs", () => {
    const copies = copyAs(
      [
        { type: "signup_login", flow: flows.signup_login_flows![1] },
        { type: "signup", flow: flows.signup_flows![1] },
        { type: "login", flow: flows.login_flows![1] },
      ],
      "members_2"
    );
    const next = applyFlowAction(flows, { kind: "add", flows: copies });
    expect(names(next, "signup_flows")).toEqual([
      "default",
      "members",
      "members_2",
    ]);
    expect(targets(next)[2]).toEqual(["members_2", "members_2"]);
  });
});

describe("delete", () => {
  it("deletes a bundle's three flows", () => {
    const next = applyFlowAction(flows, {
      kind: "delete",
      type: "signup_login",
      name: "members",
      bundle: true,
    });
    expect(names(next, "login_flows")).toEqual(["default"]);
  });

  it("blocks deleting a login flow a signup_login flow still names", () => {
    expect(deleteBlockers(flows, "login", "members", false)).toEqual([
      "members",
    ]);
    expect(deleteBlockers(flows, "signup_login", "members", true)).toEqual([]);
  });
});

describe("nameError", () => {
  it("checks the pattern, clashes across a bundle's types, and default", () => {
    const all = ["signup_login", "signup", "login"] as const;
    expect(nameError(flows, all, "2fa", false)).toEqual("invalid");
    expect(nameError(flows, all, "members", false)).toEqual("taken");
    expect(nameError(flows, all, "default", false)).toEqual("taken");
    expect(nameError({}, ["login"], "default", false)).toEqual("reserved");
    expect(nameError({}, ["login"], "default", true)).toEqual(null);
    expect(nameError(flows, all, "staff", false)).toEqual(null);
  });
});

describe("freeName", () => {
  it("numbers from 2 past the names taken in any of the types", () => {
    const f: AuthenticationFlowsConfig = {
      signup_flows: [{ name: "signup_login_2", steps: [] }],
    };
    expect(
      freeName(f, ["signup_login", "signup", "login"], "signup_login")
    ).toEqual("signup_login_3");
    expect(freeName(f, ["login"], "login")).toEqual("login_2");
  });
});
