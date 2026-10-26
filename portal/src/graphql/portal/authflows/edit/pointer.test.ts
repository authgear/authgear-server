import { describe, it, expect } from "@jest/globals";
import { optionOf } from "./pointer";

describe("optionOf", () => {
  it("is the option whose steps the point is in", () => {
    expect(optionOf({ steps: "/steps/0/one_of/1/steps", index: 0 })).toEqual({
      choice: "/steps/0",
      optionIndex: 1,
    });
    expect(
      optionOf({ steps: "/steps/0/one_of/1/steps/2/one_of/0/steps", index: 3 })
    ).toEqual({ choice: "/steps/0/one_of/1/steps/2", optionIndex: 0 });
  });

  it("is null at the top level", () => {
    expect(optionOf({ steps: "/steps", index: 1 })).toEqual(null);
  });
});
