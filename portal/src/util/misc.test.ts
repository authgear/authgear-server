/* global describe, it, expect */
import { clearEmptyObject } from "./misc";

describe("clear empty child object from object", () => {
  it("no empty object", () => {
    const TEST_OBJ = { a: "1", b: { ba: "2" } };
    clearEmptyObject(TEST_OBJ);
    expect(TEST_OBJ).toEqual(TEST_OBJ);
  });

  it("nested empty object", () => {
    const TEST_OBJ = { a: "1", b: { ba: {}, bb: "2" }, c: {}, d: { da: {} } };
    const EXPECTED_RESULT = { a: "1", b: { bb: "2" } };
    clearEmptyObject(TEST_OBJ);
    expect(TEST_OBJ).toEqual(EXPECTED_RESULT);
  });
});
