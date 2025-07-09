import { describe, it, expect } from "@jest/globals";
import { generateLabel } from "./label";

describe("generateLabel", () => {
  it("generates label", () => {
    expect(generateLabel("a")).toEqual("A");
    expect(generateLabel("a_pen")).toEqual("A Pen");
    expect(generateLabel("foobar")).toEqual("Foobar");
    expect(generateLabel("a_to_b")).toEqual("A to B");
    expect(generateLabel("a_b_c_d")).toEqual("A B C D");
    expect(generateLabel("a_b_🙂_d")).toEqual("A B 🙂 D");
  });
});
