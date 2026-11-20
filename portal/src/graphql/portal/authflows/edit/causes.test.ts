import { describe, expect, it } from "@jest/globals";
import { ValidationFailedErrorInfoCause } from "../../../../error/validation";
import { describeCause, flowErrorRule } from "./causes";

const at = "/authentication_flow/signup_flows/0/steps/2";

function general(
  location: string,
  msg: string
): ValidationFailedErrorInfoCause {
  return { location, kind: "general", details: { msg } };
}

describe("describeCause", () => {
  it("words the server's form checks", () => {
    expect(
      describeCause(
        general(at, "fill_form must come after an identify step"),
        at
      )
    ).toEqual({ messageID: "AuthFlowsScreen.check.form-before-identify" });
    expect(
      describeCause(
        general(
          `${at}/form/fields/1/user_profile/pointer`,
          "attribute is not fillable"
        ),
        at
      )
    ).toEqual({ messageID: "AuthFlowsScreen.error.not-fillable" });
    expect(
      describeCause(
        general(
          `${at}/form/fields/1/user_profile/pointer`,
          "duplicated pointer"
        ),
        at
      )
    ).toEqual({ messageID: "AuthFlowsScreen.error.duplicate-field" });
  });

  it("words a missing target", () => {
    expect(
      describeCause(
        general(
          `${at}/target_step`,
          "target_step does not refer to any named step in the same flow"
        ),
        at
      )
    ).toEqual({ messageID: "AuthFlowsScreen.error.target-missing" });
  });

  it("passes other server messages through", () => {
    expect(describeCause(general(at, "something else"), at)).toEqual({
      messageID: "AuthFlowsScreen.error.server",
      values: { message: "something else" },
    });
  });

  it("names the setting relative to the step, not the config", () => {
    expect(
      describeCause(
        {
          location: `${at}/form`,
          kind: "required",
          details: { actual: [], expected: ["fields"], missing: ["fields"] },
        },
        at
      )
    ).toEqual({
      messageID: "AuthFlowsScreen.error.required",
      values: { fields: "form.fields" },
    });
    expect(
      describeCause(
        {
          location: `${at}/form/fields`,
          kind: "minItems",
          details: { actual: 0, expected: 1 },
        },
        at
      )
    ).toEqual({
      messageID: "AuthFlowsScreen.error.min-items",
      values: { field: "form.fields", count: 1 },
    });
    expect(
      describeCause(
        {
          location: `${at}/one_of/0/identification`,
          kind: "format",
          details: { format: "x" },
        },
        at
      )
    ).toEqual({
      messageID: "AuthFlowsScreen.error.invalid",
      values: { field: "one_of.0.identification" },
    });
  });

  it("words a problem with the step itself", () => {
    expect(
      describeCause(
        { location: at, kind: "format", details: { format: "x" } },
        at
      )
    ).toEqual({ messageID: "AuthFlowsScreen.error.invalid-step" });
  });
});

describe("flowErrorRule", () => {
  const prefix = "/authentication_flow/signup_flows/0";
  const apiError = {
    errorName: "Invalid",
    reason: "ValidationFailed" as const,
    info: {
      causes: [
        general(`${prefix}/steps/0`, "a"),
        general(`${prefix}/steps/1/form`, "b"),
        general("/http/allowed_origins", "c"),
      ],
    },
  };

  it("counts the flow's causes in one line and leaves the rest", () => {
    const result = flowErrorRule(prefix)(apiError);
    expect(result.parsedAPIErrors).toEqual([
      { messageID: "AuthFlowsScreen.error.on-flow", arguments: { count: 2 } },
    ]);
    expect(
      (result.modifiedAPIError as typeof apiError).info.causes.map(
        (c) => c.location
      )
    ).toEqual(["/http/allowed_origins"]);
    expect(result.fullyHandled).toBe(false);
  });

  it("handles an error entirely inside the flow", () => {
    const inside = {
      ...apiError,
      info: { causes: apiError.info.causes.slice(0, 2) },
    };
    expect(flowErrorRule(prefix)(inside as never).fullyHandled).toBe(true);
  });

  it("leaves other errors alone", () => {
    const other = { errorName: "x", reason: "Forbidden" };
    expect(flowErrorRule(prefix)(other as never)).toEqual({
      parsedAPIErrors: [],
      fullyHandled: false,
    });
  });
});
