import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowConfig } from "../../../../types";
import {
  duplicateNames,
  earlierSteps,
  ensureNamed,
  referenceWarnings,
  referencesTo,
  removalBlockers,
  renameStep,
  targetCandidates,
} from "./references";
import {
  generatedEmailOAuthPasskey2FA,
  generatedUsernameOnly,
} from "../__fixtures__/flows";

// identify (email → verify, create password) then a top-level authenticate
// that refers back to identify.
const signup: AuthenticationFlowConfig = {
  name: "default",
  steps: [
    {
      type: "identify",
      name: "identify",
      one_of: [
        {
          identification: "email",
          steps: [
            { type: "verify", target_step: "identify" },
            {
              type: "create_authenticator",
              name: "create_password",
              one_of: [{ authentication: "primary_password" }],
            },
          ],
        },
        { identification: "phone" },
      ],
    },
    { type: "view_recovery_code" },
  ],
};

describe("earlierSteps", () => {
  it("includes ancestors and earlier siblings, not the step or later ones", () => {
    expect(earlierSteps(signup, "/steps/0/one_of/0/steps/1")).toEqual([
      "/steps/0",
      "/steps/0/one_of/0/steps/0",
    ]);
  });

  it("includes what is nested under an earlier sibling", () => {
    expect(earlierSteps(signup, "/steps/1")).toEqual([
      "/steps/0",
      "/steps/0/one_of/0/steps/0",
      "/steps/0/one_of/0/steps/1",
    ]);
  });
});

describe("targetCandidates", () => {
  it("lists earlier steps of the type the usage targets", () => {
    expect(
      targetCandidates(signup, "/steps/0/one_of/0/steps/1", "verify")
    ).toEqual([{ pointer: "/steps/0", name: "identify" }]);
  });

  it("excludes an identify whose options are all id_token for authenticate", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          name: "a",
          one_of: [{ identification: "id_token" }],
        },
        { type: "identify", name: "b", one_of: [{ identification: "email" }] },
        { type: "authenticate", one_of: [] },
      ],
    };
    expect(targetCandidates(flow, "/steps/2", "authenticate")).toEqual([
      { pointer: "/steps/1", name: "b" },
    ]);
  });

  it("keeps unnamed candidates, with a null name", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        { type: "authenticate", one_of: [] },
        { type: "change_password", target_step: "x" },
      ],
    };
    expect(targetCandidates(flow, "/steps/1", "change_password")).toEqual([
      { pointer: "/steps/0", name: null },
    ]);
  });
});

describe("referencesTo and removalBlockers", () => {
  it("finds the steps that name a step", () => {
    expect(referencesTo(signup, "/steps/0")).toEqual([
      "/steps/0/one_of/0/steps/0",
    ]);
  });

  it("allows removing a step together with every step that names it", () => {
    expect(removalBlockers(signup, "/steps/0")).toEqual([]);
  });

  it("blocks removing a targeted step referred to from outside", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [{ type: "identify", name: "inner", one_of: [] }],
            },
          ],
        },
        { type: "verify", target_step: "inner" },
      ],
    };
    expect(removalBlockers(flow, "/steps/0")).toEqual(["/steps/1"]);
    expect(removalBlockers(flow, "/steps/1")).toEqual([]);
  });

  it("does not block when another step outside has the same name", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        { type: "identify", name: "x", one_of: [] },
        { type: "identify", name: "x", one_of: [] },
        { type: "verify", target_step: "x" },
      ],
    };
    expect(removalBlockers(flow, "/steps/1")).toEqual([]);
  });
});

describe("renameStep", () => {
  it("rewrites the references that named the old name", () => {
    const r = renameStep(signup, "/steps/0", "identify_email");
    expect(r.flow.steps?.[0].name).toEqual("identify_email");
    expect(r.flow.steps?.[0].one_of?.[0].steps?.[0].target_step).toEqual(
      "identify_email"
    );
    expect(r.rewritten).toBe(true);
  });

  it("leaves references alone when the old name is shared", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        { type: "identify", name: "x", one_of: [] },
        { type: "identify", name: "x", one_of: [] },
        { type: "verify", target_step: "x" },
      ],
    };
    const r = renameStep(flow, "/steps/0", "y");
    expect(r.flow.steps?.[2].target_step).toEqual("x");
    expect(r.rewritten).toBe(false);
  });

  it("removes the name when cleared", () => {
    const r = renameStep(signup, "/steps/0/one_of/0/steps/1", "");
    expect(r.flow.steps?.[0].one_of?.[0].steps?.[1]).not.toHaveProperty("name");
  });
});

describe("ensureNamed", () => {
  it("keeps an existing name", () => {
    expect(ensureNamed(signup, "/steps/0", "identify").name).toEqual(
      "identify"
    );
  });

  it("picks a name no other step has", () => {
    const r = ensureNamed(signup, "/steps/1", "identify");
    expect(r.name).toEqual("identify_2");
    expect(r.flow.steps?.[1].name).toEqual("identify_2");
  });
});

describe("referenceWarnings", () => {
  it("finds nothing in generated flows", () => {
    for (const fixture of [
      generatedEmailOAuthPasskey2FA,
      generatedUsernameOnly,
    ]) {
      for (const f of fixture.signup_flows ?? []) {
        expect(referenceWarnings("signup", f).size).toEqual(0);
      }
      for (const f of fixture.login_flows ?? []) {
        expect(referenceWarnings("login", f).size).toEqual(0);
      }
    }
  });

  it("flags a missing, wrongly typed or later target", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        { type: "verify", target_step: "nowhere" },
        { type: "verify", target_step: "pw" },
        { type: "verify", target_step: "later" },
        {
          type: "create_authenticator",
          name: "pw",
          one_of: [{ authentication: "primary_password" }],
        },
        { type: "identify", name: "later", one_of: [] },
      ],
    };
    expect(referenceWarnings("signup", flow)).toEqual(
      new Map([
        ["/steps/0", "missing"],
        ["/steps/1", "wrong_type"],
        ["/steps/2", "not_earlier"],
      ])
    );
  });

  it("checks option references too", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "authenticate",
          one_of: [
            { authentication: "primary_oob_otp_email", target_step: "x" },
          ],
        },
      ],
    };
    expect(referenceWarnings("login", flow)).toEqual(
      new Map([["/steps/0", "missing"]])
    );
  });
});

describe("duplicateNames", () => {
  it("lists names used more than once", () => {
    expect(
      duplicateNames({
        name: "default",
        steps: [
          { type: "identify", name: "x", one_of: [] },
          { type: "identify", name: "x", one_of: [] },
          { type: "identify", name: "y", one_of: [] },
        ],
      })
    ).toEqual(new Set(["x"]));
  });
});

describe("removalBlockers for an option", () => {
  it("blocks removing an option whose nested step is targeted from outside", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [{ type: "identify", name: "pw", one_of: [] }],
            },
            { identification: "phone" },
          ],
        },
        { type: "verify", target_step: "pw" },
      ],
    };
    expect(removalBlockers(flow, "/steps/0/one_of/0")).toEqual(["/steps/1"]);
    expect(removalBlockers(flow, "/steps/0/one_of/1")).toEqual([]);
  });
});

describe("earlierSteps with a pointer past the flow", () => {
  it("returns only steps that exist", () => {
    // As a line drawn for the flow before a step was removed would ask.
    expect(earlierSteps(signup, "/steps/5/one_of/0/steps/0")).toEqual([
      "/steps/0",
      "/steps/0/one_of/0/steps/0",
      "/steps/0/one_of/0/steps/1",
      "/steps/1",
    ]);
    expect(
      targetCandidates(signup, "/steps/5/one_of/0/steps/0", "verify")
    ).toEqual([{ pointer: "/steps/0", name: "identify" }]);
  });
});
