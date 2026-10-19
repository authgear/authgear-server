import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowsConfig } from "../../../../types";
import {
  applyFlowAction,
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
  signup_login_flows: [handoff("default", "default")],
  signup_flows: [{ name: "default", steps: [] }],
  login_flows: [
    { name: "default", steps: [] },
    { name: "staff", steps: [] },
  ],
};

const options = (f: AuthenticationFlowsConfig) =>
  f.signup_login_flows?.[0].steps?.[0].one_of?.[0];

describe("rename", () => {
  it("renames one flow, and the hand-offs that named it follow", () => {
    const next = applyFlowAction(flows, {
      kind: "rename",
      type: "login",
      name: "default",
      newName: "members",
    });
    expect(next.login_flows?.map((f) => f.name)).toEqual(["members", "staff"]);
    expect(options(next)).toMatchObject({
      login_flow: "members",
      signup_flow: "default",
    });
  });

  it("leaves hand-offs alone when renaming a signup_login flow", () => {
    const next = applyFlowAction(flows, {
      kind: "rename",
      type: "signup_login",
      name: "default",
      newName: "members",
    });
    expect(next.signup_login_flows?.[0].name).toEqual("members");
    expect(options(next)).toMatchObject({
      login_flow: "default",
      signup_flow: "default",
    });
  });
});

describe("add", () => {
  it("appends the flow to its type", () => {
    const next = applyFlowAction(flows, {
      kind: "add",
      type: "login",
      flow: { name: "staff_2", steps: [] },
    });
    expect(next.login_flows?.map((f) => f.name)).toEqual([
      "default",
      "staff",
      "staff_2",
    ]);
  });
});

describe("delete", () => {
  it("deletes one flow", () => {
    const next = applyFlowAction(flows, {
      kind: "delete",
      type: "login",
      name: "staff",
    });
    expect(next.login_flows?.map((f) => f.name)).toEqual(["default"]);
  });

  it("drops a type's list when its last flow goes: the schema wants none or some", () => {
    const next = applyFlowAction(
      { signup_login_flows: [handoff("x", "default")] },
      { kind: "delete", type: "signup_login", name: "x" }
    );
    expect(next).toEqual({});
  });

  it("never blocks resetting a default flow: the generated one takes its place", () => {
    expect(deleteBlockers(flows, "login", "default")).toEqual([]);
  });

  it("blocks deleting a flow a signup_login flow still hands off to", () => {
    const f: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("default", "staff")],
      login_flows: [{ name: "staff", steps: [] }],
    };
    expect(deleteBlockers(f, "login", "staff")).toEqual(["default"]);
    expect(deleteBlockers(flows, "login", "staff")).toEqual([]);
    expect(deleteBlockers(flows, "signup_login", "default")).toEqual([]);
  });
});

describe("nameError", () => {
  it("checks the pattern, clashes within the type, and the reserved default", () => {
    expect(nameError(flows, "login", "2fa")).toEqual("invalid");
    expect(nameError(flows, "login", "staff")).toEqual("taken");
    expect(nameError(flows, "signup", "staff")).toEqual(null);
    expect(nameError({}, "login", "default")).toEqual("reserved");
  });
});

describe("freeName", () => {
  it("numbers from 2 past the names the type has", () => {
    const f: AuthenticationFlowsConfig = {
      login_flows: [{ name: "login_2", steps: [] }],
    };
    expect(freeName(f, "login", "login")).toEqual("login_3");
    expect(freeName(f, "signup", "signup")).toEqual("signup_2");
  });
});
