import { describe, it, expect } from "@jest/globals";
import { flowToTree } from "./tree";
import { findStep } from "./findStep";
import { generatedEmailOAuthPasskey2FA } from "./__fixtures__/flows";

const login = generatedEmailOAuthPasskey2FA.login_flows![0];

describe("findStep", () => {
  it("finds a step nested in branches", () => {
    const tree = flowToTree(login);
    const id = "/steps/0/one_of/1/steps/0/one_of/0/steps/1";
    expect(findStep(tree, id)?.step.type).toEqual("change_password");
  });

  it("returns null for an unknown id", () => {
    expect(findStep(flowToTree(login), "/steps/99")).toBeNull();
  });
});
