import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import { allowedStepTypes } from "../edit/rules";
import {
  STARTER_IDS,
  projectMethods,
  scratchFlows,
  starterFlows,
  starterMissing,
  stepSummary,
} from "./starters";

const config: PortalAPIAppConfig = {
  identity: { login_id: { keys: [{ type: "email" }, { type: "username" }] } },
  authentication: {
    identities: ["login_id", "oauth"],
    primary_authenticators: ["password", "oob_otp_email"],
    secondary_authenticators: ["totp"],
  },
} as PortalAPIAppConfig;

const generated = {
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
  signup_flows: [
    {
      name: "default",
      steps: [{ type: "identify", one_of: [{ identification: "email" }] }],
    },
  ],
  login_flows: [
    {
      name: "default",
      steps: [{ type: "identify", one_of: [{ identification: "email" }] }],
    },
  ],
};

describe("projectMethods", () => {
  it("reads login IDs, identities and authenticators", () => {
    expect(projectMethods(config)).toEqual({
      identifications: ["email", "username", "oauth"],
      primary: ["primary_password", "primary_oob_otp_email"],
      secondary: ["secondary_totp"],
    });
  });
});

describe("scratchFlows", () => {
  it("builds a bundle handing off to its own signup and login flows", () => {
    const out = scratchFlows("signup_login", "members", projectMethods(config));
    expect(out.map((f) => [f.type, f.flow.name])).toEqual([
      ["signup_login", "members"],
      ["signup", "members"],
      ["login", "members"],
    ]);
    expect(out[0].flow.steps?.[0].one_of?.[0]).toEqual({
      identification: "email",
      login_flow: "members",
      signup_flow: "members",
    });
    expect(out[2].flow.steps?.map((s) => s.type)).toEqual([
      "identify",
      "authenticate",
    ]);
  });
});

describe("starterMissing", () => {
  it("names what the project lacks", () => {
    const m = projectMethods(config);
    expect(starterMissing("email_password", m)).toEqual([]);
    expect(starterMissing("phone_code", m)).toEqual([
      "phone_login",
      "sms_code",
    ]);
    expect(starterMissing("current", m)).toEqual([]);
  });
});

describe("starterFlows", () => {
  it("copies the generated flows under the new name", () => {
    const out = starterFlows("current", "signup_login", "members", generated);
    expect(out[0].flow.steps?.[0].one_of?.[0]).toMatchObject({
      login_flow: "members",
      signup_flow: "members",
    });
  });

  it("uses only step types the editor allows", () => {
    for (const id of STARTER_IDS.filter((x) => x !== "current")) {
      for (const type of ["signup_login", "signup", "login"] as const) {
        for (const { type: t, flow } of starterFlows(
          id,
          type,
          "x",
          generated
        )) {
          const walk = (steps: typeof flow.steps): string[] =>
            (steps ?? []).flatMap((s) => [
              s.type,
              ...(s.one_of ?? []).flatMap((o) => walk(o.steps)),
            ]);
          for (const stepType of walk(flow.steps)) {
            expect(allowedStepTypes(t)).toContain(stepType);
          }
        }
      }
    }
  });
});

describe("stepSummary", () => {
  it("follows the first option of each choice", () => {
    const [signup] = starterFlows("email_password", "signup", "x", generated);
    expect(stepSummary(signup.flow)).toEqual([
      "identify",
      "verify",
      "create_authenticator",
    ]);
  });
});
