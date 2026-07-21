import { describe, it, expect } from "@jest/globals";
import {
  escapeMessageFormatText,
  unescapeMessageFormatText,
} from "./messageFormat";

describe("escapeMessageFormatText", () => {
  it("doubles single quotes", () => {
    expect(escapeMessageFormatText("O'Brien")).toEqual("O''Brien");
    expect(escapeMessageFormatText("Sam's App")).toEqual("Sam''s App");
    expect(escapeMessageFormatText("no quotes")).toEqual("no quotes");
    expect(escapeMessageFormatText("''")).toEqual("''''");
  });
});

describe("unescapeMessageFormatText", () => {
  it("reverses escapeMessageFormatText", () => {
    function test(text: string) {
      expect(unescapeMessageFormatText(escapeMessageFormatText(text))).toEqual(
        text
      );
    }

    test("O'Brien");
    test("Sam's App");
    test("no quotes");
    test("''");
    test("");
  });
});
