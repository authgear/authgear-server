import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowConfig } from "../../../../types";
import { planMove } from "./move";

// identify: email → [verify], phone → [verify, create pw],
// username → [create pw]; then prompt_create_passkey.
const flow: AuthenticationFlowConfig = {
  name: "default",
  steps: [
    {
      type: "identify",
      name: "identify",
      one_of: [
        {
          identification: "email",
          steps: [{ type: "verify", target_step: "identify" }],
        },
        {
          identification: "phone",
          steps: [
            { type: "verify", target_step: "identify" },
            {
              type: "create_authenticator",
              name: "phone_password",
              one_of: [{ authentication: "primary_password" }],
            },
          ],
        },
        {
          identification: "username",
          steps: [
            {
              type: "create_authenticator",
              one_of: [{ authentication: "primary_password" }],
            },
          ],
        },
      ],
    },
    { type: "prompt_create_passkey" },
  ],
};

const EMAIL = "/steps/0/one_of/0/steps";
const PHONE = "/steps/0/one_of/1/steps";
const USERNAME = "/steps/0/one_of/2/steps";
const PHONE_CREATE = `${PHONE}/1`;
const USERNAME_CREATE = `${USERNAME}/0`;

function moved(
  from: string,
  to: { steps: string; index: number },
  copy = false,
  f: AuthenticationFlowConfig = flow
) {
  const plan = planMove("signup", f, from, to, copy);
  if ("error" in plan) {
    throw new Error(`unexpected ${plan.error}`);
  }
  return plan;
}

describe("planMove", () => {
  it("moves a step into another option", () => {
    const { flow: next, selected } = moved(PHONE_CREATE, {
      steps: EMAIL,
      index: 1,
    });
    expect(next.steps?.[0].one_of?.[0].steps?.map((s) => s.type)).toEqual([
      "verify",
      "create_authenticator",
    ]);
    expect(next.steps?.[0].one_of?.[1].steps?.map((s) => s.type)).toEqual([
      "verify",
    ]);
    expect(selected).toEqual(`${EMAIL}/1`);
  });

  it("moves into an option with no steps, and drops an emptied option's steps", () => {
    const f: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            { identification: "email", steps: [{ type: "verify" }] },
            { identification: "username" },
          ],
        },
      ],
    };
    const { flow: next } = moved(
      `${EMAIL}/0`,
      { steps: "/steps/0/one_of/1/steps", index: 0 },
      false,
      f
    );
    expect(next.steps?.[0].one_of).toEqual([
      { identification: "email" },
      { identification: "username", steps: [{ type: "verify" }] },
    ]);
  });

  it("moves within one list, counting the step's own place", () => {
    const { flow: next, selected } = moved(`${PHONE}/0`, {
      steps: PHONE,
      index: 2,
    });
    expect(next.steps?.[0].one_of?.[1].steps?.map((s) => s.type)).toEqual([
      "create_authenticator",
      "verify",
    ]);
    expect(selected).toEqual(`${PHONE}/1`);
  });

  it("shares a step by moving it after the choice", () => {
    const f: AuthenticationFlowConfig = {
      ...flow,
      steps: [
        {
          ...flow.steps![0],
          one_of: flow.steps![0].one_of!.slice(0, 2),
        },
        ...flow.steps!.slice(1),
      ],
    };
    const { flow: next, selected } = moved(
      PHONE_CREATE,
      { steps: "/steps", index: 1 },
      false,
      f
    );
    expect(next.steps?.map((s) => s.type)).toEqual([
      "identify",
      "create_authenticator",
      "prompt_create_passkey",
    ]);
    expect(selected).toEqual("/steps/1");
  });

  it("keeps a later outer step's pointer right when moving out of an earlier one", () => {
    const f: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        { type: "verify" },
        {
          type: "identify",
          one_of: [
            { identification: "email", steps: [{ type: "fill_form" }] },
            { identification: "username" },
          ],
        },
      ],
    };
    const { flow: next, selected } = moved(
      "/steps/0",
      { steps: "/steps/1/one_of/0/steps", index: 0 },
      false,
      f
    );
    expect(next.steps?.map((s) => s.type)).toEqual(["identify"]);
    expect(next.steps?.[0].one_of?.[0].steps?.map((s) => s.type)).toEqual([
      "verify",
      "fill_form",
    ]);
    expect(selected).toEqual("/steps/0/one_of/0/steps/0");
  });

  it("changes nothing when a step is dropped where it is", () => {
    for (const index of [1, 2]) {
      expect(moved(PHONE_CREATE, { steps: PHONE, index }).flow).toBe(flow);
    }
  });

  it("copies a step without its name, leaving the original", () => {
    const { flow: next, selected } = moved(
      PHONE_CREATE,
      { steps: EMAIL, index: 1 },
      true
    );
    expect(next.steps?.[0].one_of?.[1].steps?.[1]).toBe(
      flow.steps?.[0].one_of?.[1].steps?.[1]
    );
    expect(next.steps?.[0].one_of?.[0].steps?.[1]).toEqual({
      type: "create_authenticator",
      one_of: [{ authentication: "primary_password" }],
    });
    expect(selected).toEqual(`${EMAIL}/1`);
  });

  it("refuses to share a step another option has its own copy of", () => {
    expect(
      planMove(
        "signup",
        flow,
        PHONE_CREATE,
        { steps: "/steps", index: 1 },
        false
      )
    ).toEqual({
      error: "other_option_has_copy",
      choice: "/steps/0",
      options: [2],
    });
    expect(
      planMove(
        "signup",
        flow,
        USERNAME_CREATE,
        { steps: "/steps", index: 2 },
        true
      )
    ).toEqual({
      error: "other_option_has_copy",
      choice: "/steps/0",
      options: [1],
    });
  });

  it("shares a 2FA step while other options only create a password", () => {
    const f: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [
                {
                  type: "create_authenticator",
                  one_of: [{ authentication: "primary_password" }],
                },
              ],
            },
            {
              identification: "phone",
              steps: [
                {
                  type: "create_authenticator",
                  one_of: [{ authentication: "primary_password" }],
                },
                {
                  type: "create_authenticator",
                  one_of: [
                    { authentication: "secondary_totp" },
                    { authentication: "secondary_oob_otp_sms" },
                  ],
                },
              ],
            },
          ],
        },
        { type: "prompt_create_passkey" },
      ],
    };
    const plan = planMove(
      "signup",
      f,
      `${PHONE}/1`,
      { steps: "/steps", index: 1 },
      false
    );
    expect("error" in plan).toEqual(false);
  });

  it("refuses to move a step into itself", () => {
    expect(
      planMove("signup", flow, "/steps/0", { steps: PHONE, index: 0 }, false)
    ).toEqual({ error: "into_itself" });
  });

  it("refuses to empty an account recovery option", () => {
    const recovery: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [{ type: "select_destination" }],
            },
            {
              identification: "phone",
              steps: [{ type: "select_destination" }],
            },
          ],
        },
      ],
    };
    expect(
      planMove(
        "account_recovery",
        recovery,
        `${EMAIL}/0`,
        { steps: PHONE, index: 1 },
        false
      )
    ).toEqual({ error: "option_needs_a_step" });
  });

  it("keeps signup_login steps at the top level", () => {
    const f: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        { type: "identify", one_of: [{ identification: "email" }] },
        { type: "identify", one_of: [{ identification: "username" }] },
      ],
    };
    expect(
      planMove(
        "signup_login",
        f,
        "/steps/1",
        { steps: "/steps/0/one_of/0/steps", index: 0 },
        false
      )
    ).toEqual({ error: "not_allowed_here" });
  });
});
