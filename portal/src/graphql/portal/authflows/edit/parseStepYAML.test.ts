import { describe, it, expect } from "@jest/globals";
import { parseStepYAML } from "./parseStepYAML";

describe("parseStepYAML", () => {
  it("parses a step of a type the flow allows, keeping every field", () => {
    expect(
      parseStepYAML(
        "type: identify\nname: x\none_of:\n  - identification: email\n    x_new: 1\n",
        "login"
      )
    ).toEqual({
      step: {
        type: "identify",
        name: "x",
        one_of: [{ identification: "email", x_new: 1 }],
      },
    });
  });

  it("refuses YAML that does not parse", () => {
    expect(parseStepYAML("type: [", "login")).toEqual({ error: "parse" });
  });

  it("refuses anything but a single step object", () => {
    expect(parseStepYAML("- type: identify", "login")).toEqual({
      error: "not_step",
    });
    expect(parseStepYAML("name: x", "login")).toEqual({ error: "not_step" });
    expect(parseStepYAML("", "login")).toEqual({ error: "not_step" });
  });

  it("refuses a type the flow type does not allow", () => {
    expect(parseStepYAML("type: verify", "login")).toEqual({
      error: "type",
    });
  });

  it("accepts the deprecated fill_in_user_profile in signup", () => {
    expect(
      parseStepYAML("type: fill_in_user_profile\nuser_profile: []", "signup")
    ).toEqual({ step: { type: "fill_in_user_profile", user_profile: [] } });
  });
});
