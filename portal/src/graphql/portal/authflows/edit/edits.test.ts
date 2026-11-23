import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowConfig } from "../../../../types";
import { getAt, parentSteps, parsePointer } from "./pointer";
import {
  addOption,
  countNestedSteps,
  insertStep,
  moveItem,
  removeOption,
  removeStep,
  replaceStep,
  updateStep,
} from "./edits";

const flow: AuthenticationFlowConfig = {
  name: "default",
  steps: [
    {
      type: "identify",
      name: "identify",
      one_of: [
        {
          identification: "email",
          steps: [{ type: "authenticate", one_of: [] }],
        },
        { identification: "phone" },
      ],
    },
    { type: "check_account_status", x_unknown: 1 } as never,
  ],
};

describe("pointer", () => {
  it("parses array indexes as numbers", () => {
    expect(parsePointer("/steps/0/one_of/1/steps")).toEqual([
      "steps",
      0,
      "one_of",
      1,
      "steps",
    ]);
  });

  it("gets a nested value, or undefined past the end", () => {
    expect(getAt(flow, "/steps/0/one_of/0/steps/0")).toEqual({
      type: "authenticate",
      one_of: [],
    });
    expect(getAt(flow, "/steps/0/one_of/1/steps")).toBeUndefined();
  });

  it("splits a step pointer into its steps array and index", () => {
    expect(parentSteps("/steps/0/one_of/1/steps/2")).toEqual({
      steps: "/steps/0/one_of/1/steps",
      index: 2,
    });
  });
});

describe("insertStep", () => {
  it("inserts at an index and selects the new step", () => {
    const r = insertStep(
      flow,
      { steps: "/steps", index: 1 },
      { type: "terminate_other_sessions" }
    );
    expect(r.flow.steps?.map((s) => s.type)).toEqual([
      "identify",
      "terminate_other_sessions",
      "check_account_status",
    ]);
    expect(r.selected).toEqual("/steps/1");
  });

  it("creates a missing option steps array", () => {
    const r = insertStep(
      flow,
      { steps: "/steps/0/one_of/1/steps", index: 0 },
      { type: "authenticate", one_of: [] }
    );
    expect(r.flow.steps?.[0].one_of?.[1].steps).toEqual([
      { type: "authenticate", one_of: [] },
    ]);
    expect(r.selected).toEqual("/steps/0/one_of/1/steps/0");
  });

  it("leaves the input and untouched branches as they were", () => {
    const r = insertStep(
      flow,
      { steps: "/steps/0/one_of/1/steps", index: 0 },
      { type: "authenticate", one_of: [] }
    );
    expect(flow.steps?.[0].one_of?.[1].steps).toBeUndefined();
    expect(r.flow.steps?.[1]).toBe(flow.steps?.[1]);
    expect(r.flow.steps?.[0].one_of?.[0]).toBe(flow.steps?.[0].one_of?.[0]);
  });
});

describe("removeStep", () => {
  it("removes the step and what is nested under it", () => {
    const r = removeStep(flow, "/steps/0");
    expect(r.flow.steps).toEqual([flow.steps?.[1]]);
    expect(r.selected).toBeNull();
  });
});

describe("options", () => {
  it("adds an option and keeps the step selected", () => {
    const r = addOption(flow, "/steps/0", { identification: "username" });
    expect(r.flow.steps?.[0].one_of?.map((o) => o.identification)).toEqual([
      "email",
      "phone",
      "username",
    ]);
    expect(r.selected).toEqual("/steps/0");
  });

  it("removes an option by index", () => {
    const r = removeOption(flow, "/steps/0", 0);
    expect(r.flow.steps?.[0].one_of).toEqual([{ identification: "phone" }]);
  });
});

describe("updateStep and replaceStep", () => {
  it("changes one field and keeps fields it does not touch", () => {
    const r = updateStep(flow, "/steps/1", (s) => {
      s.name = "status";
    });
    expect(r.flow.steps?.[1]).toEqual({
      type: "check_account_status",
      x_unknown: 1,
      name: "status",
    });
    expect(r.selected).toEqual("/steps/1");
  });

  it("replaces a whole step", () => {
    const r = replaceStep(flow, "/steps/1", { type: "change_password" });
    expect(r.flow.steps?.[1]).toEqual({ type: "change_password" });
  });
});

describe("countNestedSteps", () => {
  it("counts every step under the options, at any depth", () => {
    expect(countNestedSteps(flow.steps![0])).toEqual(1);
    expect(countNestedSteps(flow.steps![1])).toEqual(0);
  });
});

describe("removeStep in an option", () => {
  it("drops the option's steps key when its last step goes", () => {
    const r = removeStep(flow, "/steps/0/one_of/0/steps/0");
    expect(r.flow.steps?.[0].one_of?.[0]).toEqual({ identification: "email" });
  });
});

describe("moveItem", () => {
  it("moves an item up and down", () => {
    expect(moveItem(["a", "b", "c"], 2, 1)).toEqual(["a", "c", "b"]);
    expect(moveItem(["a", "b", "c"], 0, 1)).toEqual(["b", "a", "c"]);
  });

  it("changes nothing for a move out of range", () => {
    const list = ["a", "b"];
    expect(moveItem(list, 0, -1)).toEqual(["a", "b"]);
    expect(moveItem(list, 1, 2)).toEqual(["a", "b"]);
    expect(list).toEqual(["a", "b"]);
  });
});
