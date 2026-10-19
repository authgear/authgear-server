import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowsConfig } from "../../../../types";
import { flowSections, isBundle } from "./rows";

const generated: AuthenticationFlowsConfig = {
  signup_login_flows: [
    {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              login_flow: "default",
              signup_flow: "default",
            },
          ],
        },
      ],
    },
  ],
  signup_flows: [{ name: "default", steps: [{ type: "identify" }] }],
  login_flows: [{ name: "default", steps: [{ type: "identify" }] }],
  account_recovery_flows: [{ name: "default", steps: [{ type: "identify" }] }],
};

const handoff = (name: string, login: string, signup: string) => ({
  name,
  steps: [
    {
      type: "identify",
      one_of: [
        { identification: "email", login_flow: login, signup_flow: signup },
        { identification: "passkey", login_flow: login },
      ],
    },
  ],
});

describe("flowSections", () => {
  it("lists the generated defaults when there are no custom flows", () => {
    const sections = flowSections(undefined, generated);
    expect(sections.map((s) => s.key)).toEqual([
      "signup_login",
      "signup",
      "login",
      "other",
    ]);
    expect(sections[0].rows).toEqual([
      {
        type: "signup_login",
        name: "default",
        generated: true,
        inUse: true,
        bundle: true,
        partOf: null,
      },
    ]);
    expect(sections[1].rows[0]).toMatchObject({
      generated: true,
      partOf: "default",
    });
    expect(sections[3].rows.map((r) => r.type)).toEqual(["account_recovery"]);
  });

  it("puts a custom default in place of the generated one", () => {
    const custom: AuthenticationFlowsConfig = {
      login_flows: [
        { name: "default", steps: [] },
        { name: "staff", steps: [] },
      ],
    };
    const login = flowSections(custom, generated)[2].rows;
    expect(login.map((r) => [r.name, r.generated, r.inUse])).toEqual([
      ["default", false, true],
      ["staff", false, false],
    ]);
  });

  it("bundles a signup_login flow with same-named signup and login flows", () => {
    const custom: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("members", "members", "members")],
      signup_flows: [{ name: "members", steps: [] }],
      login_flows: [{ name: "members", steps: [] }],
    };
    const [sl, signup, login] = flowSections(custom, generated);
    expect(sl.rows.find((r) => r.name === "members")).toMatchObject({
      bundle: true,
    });
    expect(signup.rows.find((r) => r.name === "members")?.partOf).toEqual(
      "members"
    );
    expect(login.rows.find((r) => r.name === "members")?.partOf).toEqual(
      "members"
    );
  });

  it("does not bundle a signup_login flow that hands off to other names", () => {
    const custom: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("members", "default", "members")],
      signup_flows: [{ name: "members", steps: [] }],
      login_flows: [{ name: "members", steps: [] }],
    };
    expect(isBundle(custom, "members")).toEqual(false);
    const signup = flowSections(custom, generated)[1];
    expect(signup.rows.find((r) => r.name === "members")?.partOf).toEqual(null);
  });

  it("omits Other when the project has none", () => {
    const keys = flowSections(undefined, {
      ...generated,
      account_recovery_flows: [],
    }).map((s) => s.key);
    expect(keys).toEqual(["signup_login", "signup", "login"]);
  });
});
