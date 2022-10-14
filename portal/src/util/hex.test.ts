import { describe, it, expect } from "@jest/globals";
import { convertToHexstring, parseHexstring } from "./hex";

describe("hexstring", () => {
  it("parses hexstring", () => {
    function test(hex: string, expected: string) {
      const h = parseHexstring(hex);

      expect(h).toEqual(expected);
    }

    test(
      "0xec7f0e0c2b7a356b5271d13e75004705977fd0100000000000000300000186a0",
      "106970318795424639811796305122058490727656953019737735212233586537683535103648"
    );

    test("", "");

    test("foobar", "");
  });

  it("convert to hexstring", () => {
    function test(dec: string, expected: string) {
      const hex = convertToHexstring(dec);

      expect(hex).toEqual(expected);
    }

    test("", "");

    test("foobar", "");
  });
});
