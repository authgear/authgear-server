import { describe, it, expect } from "@jest/globals";
import { checkNonNegativeIntegerInput } from "./input";

describe("checkNonNegativeIntegerInput", () => {
  it("accepts empty and whole numbers from 0", () => {
    for (const v of ["", "0", "3", "120"]) {
      expect(checkNonNegativeIntegerInput(v)).toBe(true);
    }
  });

  it("refuses a sign, a fraction and a leading zero", () => {
    for (const v of ["-", "-1", "1.5", "01", "a"]) {
      expect(checkNonNegativeIntegerInput(v)).toBe(false);
    }
  });
});
