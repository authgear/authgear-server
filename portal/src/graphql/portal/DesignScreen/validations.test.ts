import { describe, it, expect } from "@jest/globals";
import { validateBorderRadius } from "./validations";

describe("validateBorderRadius", () => {
  it("reject invalid format", () => {
    const errs = validateBorderRadius("test", {
      type: "rounded",
      radius: "invalidstr",
    });
    expect(errs.length).toEqual(1);
    expect(errs[0].messageID).toEqual("errors.validation.borderRadius.format");
  });

  it("reject invalid unit", () => {
    const errs = validateBorderRadius("test", {
      type: "rounded",
      radius: "1meter",
    });
    expect(errs.length).toEqual(1);
    expect(errs[0].messageID).toEqual("errors.validation.borderRadius.unit");
  });

  it("reject invalid number", () => {
    const errs = validateBorderRadius("test", {
      type: "rounded",
      radius: "0.0.1rem",
    });
    expect(errs.length).toEqual(1);
    expect(errs[0].messageID).toEqual("errors.validation.borderRadius.format");
  });

  it("accept valid value", () => {
    const errs = validateBorderRadius("test", {
      type: "rounded",
      radius: "0.875em",
    });
    expect(errs.length).toEqual(0);
  });

  it("accept omitted unit for 0", () => {
    const errs = validateBorderRadius("test", { type: "rounded", radius: "0" });
    expect(errs.length).toEqual(0);
  });
});
