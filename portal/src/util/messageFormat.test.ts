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

  it("quotes the syntax characters", () => {
    expect(escapeMessageFormatText("My {App}")).toEqual("My '{App}'");
    expect(escapeMessageFormatText("{")).toEqual("'{'");
    expect(escapeMessageFormatText("}")).toEqual("'}'");
    expect(escapeMessageFormatText("{0} and {1}")).toEqual("'{0} and {1}'");
  });

  it("keeps adjacent syntax characters in one quoted run", () => {
    // Quoting each character separately would produce "'{''}'", whose two
    // inner quotes fuse into a literal apostrophe and render as "{'}".
    expect(escapeMessageFormatText("{}")).toEqual("'{}'");
    expect(escapeMessageFormatText("}{")).toEqual("'}{'");
  });

  it("handles apostrophes adjacent to syntax characters", () => {
    expect(escapeMessageFormatText("{'}")).toEqual("'{''}'");
    expect(escapeMessageFormatText("Acme '{Corp}'")).toEqual(
      "Acme '''{Corp}'''"
    );
  });

  it("leaves # alone, since these values are whole messages", () => {
    expect(escapeMessageFormatText("50% off #1")).toEqual("50% off #1");
  });

  it("never emits an unterminated quoted run", () => {
    // A stray ' would make the server reject the whole message.
    for (const text of ["'", "''", "a'", "'a", "{'", "'}"]) {
      const escaped = escapeMessageFormatText(text);
      let quoted = false;
      for (let i = 0; i < escaped.length; i++) {
        if (escaped[i] !== "'") continue;
        if (escaped[i + 1] === "'") {
          i += 1;
          continue;
        }
        quoted = !quoted;
      }
      expect(quoted).toBe(false);
    }
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
    test("My {App}");
    test("{}");
    test("}{");
    test("{'}");
    test("Acme '{Corp}'");
    test("50% off #1");
    test("Ünïcødé {ñ}");
  });

  it("round-trips exhaustively over the tricky characters", () => {
    const alphabet = ["'", "{", "}", "a", "#"];
    let cases: string[] = [""];
    for (let n = 0; n < 5; n++) {
      const next: string[] = [];
      for (const prefix of cases) {
        for (const ch of alphabet) {
          next.push(prefix + ch);
        }
      }
      for (const text of next) {
        expect(
          unescapeMessageFormatText(escapeMessageFormatText(text))
        ).toEqual(text);
      }
      cases = next;
    }
  });

  it("still reads values written by the previous escape", () => {
    // Older versions doubled apostrophes but left { and } bare.
    expect(unescapeMessageFormatText("Sam''s App")).toEqual("Sam's App");
    expect(unescapeMessageFormatText("My {App}")).toEqual("My {App}");
    expect(unescapeMessageFormatText("no quotes")).toEqual("no quotes");
  });
});
