import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowsConfig } from "../../../../types";
import { flowSections } from "./rows";

const generated: AuthenticationFlowsConfig = {
  signup_login_flows: [{ name: "default", steps: [] }],
  signup_flows: [{ name: "default", steps: [] }],
  login_flows: [{ name: "default", steps: [] }],
  account_recovery_flows: [{ name: "default", steps: [] }],
};

const names = (custom: AuthenticationFlowsConfig | undefined) =>
  flowSections(custom, generated).map((s) => [
    s.key,
    s.rows.map((r) => `${r.name}${r.generated ? " (generated)" : ""}`),
  ]);

describe("flowSections", () => {
  it("lists the generated defaults of the three creatable types only", () => {
    expect(names(undefined)).toEqual([
      ["signup_login", ["default (generated)"]],
      ["signup", ["default (generated)"]],
      ["login", ["default (generated)"]],
    ]);
  });

  it("puts a custom default in place of the generated one", () => {
    const custom: AuthenticationFlowsConfig = {
      login_flows: [
        { name: "default", steps: [] },
        { name: "staff", steps: [] },
      ],
    };
    expect(names(custom)[2]).toEqual(["login", ["default", "staff"]]);
  });

  it("lists custom flows after the generated default of their type", () => {
    const custom: AuthenticationFlowsConfig = {
      signup_flows: [{ name: "members", steps: [] }],
    };
    expect(names(custom)[1]).toEqual([
      "signup",
      ["default (generated)", "members"],
    ]);
  });

  it("puts default first, even when written after other flows", () => {
    const custom: AuthenticationFlowsConfig = {
      login_flows: [
        { name: "staff", steps: [] },
        { name: "default", steps: [] },
      ],
    };
    expect(names(custom)[2]).toEqual(["login", ["default", "staff"]]);
  });

  it("marks the flow named default", () => {
    const rows = flowSections(
      { login_flows: [{ name: "staff", steps: [] }] },
      generated
    )[2].rows;
    expect(rows.map((r) => [r.name, r.isDefault])).toEqual([
      ["default", true],
      ["staff", false],
    ]);
  });
});

describe("display names", () => {
  it("carries a custom flow's display name", () => {
    const [login] = flowSections(
      {
        login_flows: [
          { name: "staff", display_name: "Staff login", steps: [] },
        ],
      },
      {}
    );
    expect(login.rows[0].displayName).toBe("Staff login");
  });
});
