import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import { allowedStepTypes } from "../edit/rules";
import {
  STARTER_IDS,
  projectMethods,
  scratchFlows,
  scratchMissing,
  starterFlows,
  starterMissing,
  NewFlows,
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
  it("builds one signup_login flow handing off to the default flows", () => {
    const out = scratchFlows("signup_login", "members", projectMethods(config));
    expect(out).toEqual([
      {
        type: "signup_login",
        flow: {
          name: "members",
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
      },
    ]);
  });

  it("builds a login flow that identifies then authenticates", () => {
    expect(
      scratchFlows("login", "x", projectMethods(config))[0].flow.steps?.map(
        (s) => s.type
      )
    ).toEqual(["identify", "authenticate"]);
  });
});

const withPhone: PortalAPIAppConfig = {
  identity: { login_id: { keys: [{ type: "email" }, { type: "phone" }] } },
  authentication: {
    identities: ["login_id", "oauth"],
    primary_authenticators: ["password", "oob_otp_email", "oob_otp_sms"],
    secondary_authenticators: ["totp"],
  },
} as PortalAPIAppConfig;

const none = { identifications: [], primary: [], secondary: [] };

describe("starterMissing", () => {
  it("names what the project lacks", () => {
    const m = projectMethods(config);
    expect(starterMissing("current", m)).toEqual([]);
    expect(starterMissing("otp", m)).toEqual([]);
    expect(starterMissing("password", m)).toEqual([]);
    expect(starterMissing("social", m)).toEqual([]);
    expect(starterMissing("password_2fa", m)).toEqual([]);
    expect(starterMissing("verified_email_phone", m)).toEqual([
      "phone_login",
      "sms_code",
    ]);
  });

  it("asks for any one of the methods a template can use", () => {
    expect(starterMissing("otp", none)).toEqual(["otp_login"]);
    expect(starterMissing("password", none)).toEqual([
      "email_or_phone",
      "password",
    ]);
    expect(starterMissing("social", none)).toEqual(["social"]);
  });
});

const steps = (flows: NewFlows, type: string) =>
  flows.find((f) => f.type === type)?.flow.steps ?? [];

describe("starterFlows", () => {
  it("copies the generated flow under the new name, hand-offs unchanged", () => {
    const out = starterFlows(
      "current",
      "signup_login",
      "members",
      generated,
      projectMethods(config)
    );
    expect(out.map((f) => [f.type, f.flow.name])).toEqual([
      ["signup_login", "members"],
    ]);
    expect(out[0].flow.steps?.[0].one_of?.[0]).toMatchObject({
      login_flow: "default",
      signup_flow: "default",
    });
  });

  it("offers OTP codes for the login IDs that have them", () => {
    const login = steps(
      starterFlows("otp", "login", "x", generated, projectMethods(withPhone)),
      "login"
    );
    expect(login[0].one_of?.map((o) => o.identification)).toEqual([
      "email",
      "phone",
    ]);
    expect(login[0].one_of?.[1].steps?.[0].one_of).toEqual([
      { authentication: "primary_oob_otp_sms", target_step: "identify" },
    ]);
  });

  it("logs in with a password, or an OTP code when the project has it", () => {
    const login = steps(
      starterFlows("password", "login", "x", generated, projectMethods(config)),
      "login"
    );
    expect(login[0].one_of?.[0].steps?.[0].one_of).toEqual([
      { authentication: "primary_password" },
      { authentication: "primary_oob_otp_email", target_step: "identify" },
    ]);
  });

  it("verifies both email and phone at signup", () => {
    const signup = steps(
      starterFlows(
        "verified_email_phone",
        "signup",
        "x",
        generated,
        projectMethods(withPhone)
      ),
      "signup"
    );
    expect(signup.map((s) => [s.type, s.one_of?.[0].identification])).toEqual([
      ["identify", "email"],
      ["identify", "phone"],
    ]);
  });

  it("makes a signup & login template's signup and login flows too", () => {
    const out = starterFlows(
      "password",
      "signup_login",
      "members",
      generated,
      projectMethods(config)
    );
    expect(out.map((f) => [f.type, f.flow.name])).toEqual([
      ["signup_login", "members"],
      ["signup", "members"],
      ["login", "members"],
    ]);
    expect(out[0].flow.steps?.[0].one_of?.[0]).toMatchObject({
      identification: "email",
      login_flow: "members",
      signup_flow: "members",
    });
  });

  it("uses only step types the editor allows", () => {
    const m = projectMethods(withPhone);
    for (const id of STARTER_IDS.filter((x) => x !== "current")) {
      for (const type of ["signup_login", "signup", "login"] as const) {
        for (const { type: t, flow } of starterFlows(
          id,
          type,
          "x",
          generated,
          m
        )) {
          const walk = (s: typeof flow.steps): string[] =>
            (s ?? []).flatMap((x) => [
              x.type,
              ...(x.one_of ?? []).flatMap((o) => walk(o.steps)),
            ]);
          for (const stepType of walk(flow.steps)) {
            expect(allowedStepTypes(t)).toContain(stepType);
          }
        }
      }
    }
  });
});

describe("scratchFlows with few methods", () => {
  it("prefers a login ID over social login", () => {
    const m = {
      identifications: ["oauth", "phone"],
      primary: ["primary_oob_otp_sms"],
      secondary: [],
    };
    expect(scratchFlows("login", "x", m)[0].flow.steps?.[0].one_of).toEqual([
      { identification: "phone" },
    ]);
  });

  it("skips the authenticate step after social login", () => {
    const m = { identifications: ["oauth"], primary: [], secondary: [] };
    expect(
      scratchFlows("login", "x", m)[0].flow.steps?.map((s) => s.type)
    ).toEqual(["identify"]);
  });

  it("is unavailable without a login ID or social login", () => {
    expect(
      scratchMissing({
        identifications: ["passkey"],
        primary: [],
        secondary: [],
      })
    ).toEqual(true);
    expect(
      scratchMissing({ identifications: ["email"], primary: [], secondary: [] })
    ).toEqual(false);
  });
});
