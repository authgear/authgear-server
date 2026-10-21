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

  it("refuses a form where the server does not allow one", () => {
    expect(
      parseStepYAML("type: authenticate\nform: {fields: []}", "login")
    ).toEqual({ error: "form" });
    expect(
      parseStepYAML(
        "type: identify\none_of:\n- identification: email\n  form: {fields: []}",
        "login"
      )
    ).toEqual({ error: "form" });
    expect(
      parseStepYAML(
        "type: identify\none_of:\n- identification: oauth\n  form: {fields: []}",
        "signup"
      )
    ).toEqual({ error: "form" });
    expect(
      parseStepYAML(
        "type: identify\none_of:\n- identification: email\n  steps:\n  - type: verify\n    form: {fields: []}",
        "signup"
      )
    ).toEqual({ error: "form" });
  });

  it("accepts a form on a fill_form step and a signup email option", () => {
    expect(
      "step" in parseStepYAML("type: fill_form\nform: {fields: []}", "login")
    ).toBe(true);
    expect(
      "step" in
        parseStepYAML(
          "type: identify\none_of:\n- identification: email\n  form: {fields: []}",
          "signup"
        )
    ).toBe(true);
  });
});
