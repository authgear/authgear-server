import { describe, it, expect } from "@jest/globals";
import { flowToTree, FlowStepItem } from "./tree";
import {
  customSignupWithForm,
  generatedEmailOAuthPasskey2FA,
  generatedUsernameOnly,
} from "./__fixtures__/flows";
import { AuthenticationFlowConfig } from "../../../types";

function only<T>(xs: T[] | undefined): T {
  if (xs?.length !== 1) {
    throw new Error("expected exactly one element");
  }
  return xs[0];
}

function summary(items: FlowStepItem[]) {
  return items.map((i) => ({
    id: i.id,
    kind: i.kind,
    branches: i.branches?.length ?? null,
  }));
}

describe("flowToTree", () => {
  it("builds the generated login flow", () => {
    const tree = flowToTree(only(generatedEmailOAuthPasskey2FA.login_flows));
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "page", branches: 4 },
      { id: "/steps/1", kind: "pageless", branches: null },
      // terminate_other_sessions asks the user to confirm on a page.
      { id: "/steps/2", kind: "page", branches: null },
      { id: "/steps/3", kind: "page", branches: null },
    ]);
    const emailBranch = tree[0].branches![1];
    expect(summary(emailBranch.sequence)).toEqual([
      { id: "/steps/0/one_of/1/steps/0", kind: "page", branches: 2 },
    ]);
    const passwordBranch = emailBranch.sequence[0].branches![0];
    expect(passwordBranch.sequence[0].optional).toBe(true);
    expect(summary(passwordBranch.sequence)).toEqual([
      {
        id: "/steps/0/one_of/1/steps/0/one_of/0/steps/0",
        kind: "page",
        // device_token is not a choice: it lets a remembered device skip
        // the step, so it does not become a branch.
        branches: 3,
      },
      {
        id: "/steps/0/one_of/1/steps/0/one_of/0/steps/1",
        kind: "page",
        branches: null,
      },
    ]);
  });

  it("inlines single-option steps", () => {
    const tree = flowToTree(only(customSignupWithForm.signup_flows));
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "page", branches: null },
      { id: "/steps/0/one_of/0/steps/0", kind: "page", branches: null },
      { id: "/steps/0/one_of/0/steps/1", kind: "page", branches: null },
      { id: "/steps/1", kind: "page", branches: null },
    ]);
  });

  it("treats identify by id_token as pageless", () => {
    const tree = flowToTree(only(generatedEmailOAuthPasskey2FA.reauth_flows));
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "pageless", branches: null },
      { id: "/steps/1", kind: "page", branches: 4 },
    ]);
  });

  it("keeps a choice step without options as an empty choice", () => {
    const tree = flowToTree(only(generatedUsernameOnly.account_recovery_flows));
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "page", branches: 0 },
      { id: "/steps/1", kind: "page", branches: null },
      { id: "/steps/2", kind: "page", branches: null },
    ]);
  });

  it("does not inline the single option of an optional step", () => {
    const flow: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "authenticate",
          optional: true,
          one_of: [
            {
              authentication: "secondary_totp",
              steps: [{ type: "view_recovery_code" }],
            },
          ],
        },
      ],
    };
    const tree = flowToTree(flow);
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "page", branches: 1 },
    ]);
    expect(summary(tree[0].branches![0].sequence)).toEqual([
      { id: "/steps/0/one_of/0/steps/0", kind: "page", branches: null },
    ]);
  });

  it("keeps the original option index on branches after device_token", () => {
    const tree = flowToTree({
      name: "default",
      steps: [
        {
          type: "authenticate",
          one_of: [
            { authentication: "device_token" },
            { authentication: "secondary_totp" },
            { authentication: "recovery_code" },
          ],
        },
      ],
    });
    expect(tree[0].branches?.map((b) => b.optionIndex)).toEqual([1, 2]);
  });

  it("inlines a step whose only other option is device_token", () => {
    const tree = flowToTree({
      name: "default",
      steps: [
        {
          type: "authenticate",
          one_of: [
            { authentication: "secondary_totp" },
            { authentication: "device_token" },
          ],
        },
      ],
    });
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "page", branches: null },
    ]);
  });

  it("draws steps every option shares once, after the options rejoin", () => {
    const recovery = { type: "view_recovery_code" };
    const tree = flowToTree({
      name: "default",
      steps: [
        {
          type: "create_authenticator",
          one_of: [
            { authentication: "secondary_totp", steps: [recovery] },
            { authentication: "secondary_oob_otp_sms", steps: [recovery] },
          ],
        },
        { type: "prompt_create_passkey" },
      ],
    });
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "page", branches: 2 },
      { id: "/steps/0/one_of/0/steps/0", kind: "page", branches: null },
      { id: "/steps/1", kind: "page", branches: null },
    ]);
    expect(tree[0].branches?.map((b) => b.sequence.length)).toEqual([0, 0]);
  });

  it("keeps branches apart when their steps differ", () => {
    const tree = flowToTree({
      name: "default",
      steps: [
        {
          type: "create_authenticator",
          one_of: [
            {
              authentication: "secondary_totp",
              steps: [{ type: "view_recovery_code" }],
            },
            { authentication: "secondary_oob_otp_sms" },
          ],
        },
      ],
    });
    expect(tree[0].branches?.map((b) => b.sequence.length)).toEqual([1, 0]);
  });

  it("marks unknown step types", () => {
    const tree = flowToTree({
      name: "default",
      steps: [{ type: "future_step" }],
    });
    expect(summary(tree)).toEqual([
      { id: "/steps/0", kind: "unknown", branches: null },
    ]);
  });
});
