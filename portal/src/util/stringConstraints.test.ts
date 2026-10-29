import { describe, it, expect } from "@jest/globals";
import {
  constraintsFromDraft,
  draftFromConstraints,
  applyValidationUpdates,
  hasConstraints,
  patternErrorKey,
  summaryParts,
} from "./stringConstraints";

describe("constraintsFromDraft", () => {
  it("writes only what is set", () => {
    expect(
      constraintsFromDraft({
        ...draftFromConstraints(undefined),
        minLength: "3",
        maxLength: "",
      })
    ).toEqual({ min_length: 3 });
  });

  it("drops allowed_characters for Any, and for Only with nothing ticked", () => {
    const d = draftFromConstraints({ allowed_characters: ["letters"] });
    expect(constraintsFromDraft(d)).toEqual({
      allowed_characters: ["letters"],
    });
    expect(constraintsFromDraft({ ...d, charactersLimited: false })).toEqual(
      {}
    );
    expect(constraintsFromDraft({ ...d, allowedCharacters: [] })).toEqual({});
  });

  it("keeps the characters in a fixed order", () => {
    const d = {
      ...draftFromConstraints(undefined),
      charactersLimited: true,
      allowedCharacters: ["spaces", "letters"] as const,
    };
    expect(
      constraintsFromDraft({
        ...d,
        allowedCharacters: [...d.allowedCharacters],
      })
    ).toEqual({
      allowed_characters: ["letters", "spaces"],
    });
  });

  it("round-trips a full set of constraints", () => {
    const c = {
      min_length: 2,
      max_length: 10,
      allowed_characters: ["letters" as const, "digits" as const],
      pattern: "^A",
    };
    expect(constraintsFromDraft(draftFromConstraints(c))).toEqual(c);
  });
});

describe("hasConstraints", () => {
  it("is false for an empty set", () => {
    expect(hasConstraints({})).toBe(false);
    expect(hasConstraints({ pattern: "x" })).toBe(true);
  });
});

describe("summaryParts", () => {
  it("lists length, characters and pattern", () => {
    expect(
      summaryParts({
        min_length: 3,
        max_length: 20,
        allowed_characters: ["letters", "spaces"],
        pattern: "x",
      })
    ).toEqual([
      { kind: "length-range", min: 3, max: 20 },
      { kind: "characters", characters: ["letters", "spaces"] },
      { kind: "pattern" },
    ]);
    expect(summaryParts({ max_length: 5 })).toEqual([
      { kind: "length-max", max: 5 },
    ]);
    expect(summaryParts({ min_length: 5 })).toEqual([
      { kind: "length-min", min: 5 },
    ]);
  });
});

describe("patternErrorKey", () => {
  it("names the translation key of a pointer", () => {
    expect(patternErrorKey("custom", "/x_employee_id")).toBe(
      "custom-attribute-pattern-error-/x_employee_id"
    );
    expect(patternErrorKey("standard", "/given_name")).toBe(
      "standard-attribute-pattern-error-/given_name"
    );
  });
});

describe("applyValidationUpdates", () => {
  it("replaces an entry in place, adds a new one, and drops an emptied one", () => {
    const list = [
      { pointer: "/name", max_length: 50 },
      { pointer: "/nickname", min_length: 2 },
    ];
    expect(
      applyValidationUpdates(list, [
        { pointer: "/name", constraints: { max_length: 30 } },
        { pointer: "/nickname", constraints: {} },
        { pointer: "/given_name", constraints: { pattern: "^A" } },
      ])
    ).toEqual([
      { pointer: "/name", max_length: 30 },
      { pointer: "/given_name", pattern: "^A" },
    ]);
  });
});
