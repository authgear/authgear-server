import { describe, it, expect } from "@jest/globals";
import { nullishCoalesce, or_ } from "./operators";

describe("or_", () => {
  it("should work like || operator", () => {
    const exprs0 = [true, false];
    const exprs1 = [true, false];
    const exprs2 = [true, false];

    for (const e0 of exprs0) {
      for (const e1 of exprs1) {
        for (const e2 of exprs2) {
          expect(or_(e0, e1, e2)).toEqual(e0 || e1 || e2);
        }
      }
    }
  });
});

describe("nullishCoalesce", () => {
  it("should work like ?? operator", () => {
    const exprs0 = [1, null];
    const exprs1 = [2, null];
    const exprs2 = [3, null];

    for (const e0 of exprs0) {
      for (const e1 of exprs1) {
        for (const e2 of exprs2) {
          expect(nullishCoalesce(e0, e1, e2)).toEqual(e0 ?? e1 ?? e2);
        }
      }
    }
  });
});
