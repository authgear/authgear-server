/* global describe, it, expect */
import { generateLabel } from "./label";

describe("generateLabel", () => {
  it("generates label", () => {
    expect(generateLabel("a")).toEqual("A");
    expect(generateLabel("a_pen")).toEqual("A Pen");
    expect(generateLabel("foobar")).toEqual("Foobar");
    expect(generateLabel("a_to_b")).toEqual("A to B");
  });
});
