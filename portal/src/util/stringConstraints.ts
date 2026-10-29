// String constraints on attributes (docs/specs/user-profile/design.md#string-constraints):
// on a custom attribute of type string, and in
// user_profile.standard_attributes.validation.

import {
  AttributeAllowedCharacter,
  AttributeStringConstraints,
  StandardAttributesValidationConfig,
} from "../types";

export type AllowedCharacter = AttributeAllowedCharacter;
export type StringConstraints = AttributeStringConstraints;

export const ALLOWED_CHARACTERS: readonly AllowedCharacter[] = [
  "letters",
  "digits",
  "spaces",
  "symbols",
];

// The standard attributes the schema lets constraints be set on.
export const STANDARD_VALIDATION_POINTERS = [
  "/name",
  "/given_name",
  "/family_name",
  "/middle_name",
  "/nickname",
  "/address/formatted",
  "/address/street_address",
  "/address/locality",
  "/address/region",
  "/address/postal_code",
] as const;
export type StandardValidationPointer =
  (typeof STANDARD_VALIDATION_POINTERS)[number];

export const ADDRESS_VALIDATION_POINTERS: readonly StandardValidationPointer[] =
  STANDARD_VALIDATION_POINTERS.filter((p) => p.startsWith("/address/"));

export interface StringConstraintsDraft {
  minLength: string;
  maxLength: string;
  // "Only these characters"; false is any character.
  charactersLimited: boolean;
  allowedCharacters: AllowedCharacter[];
  pattern: string;
}

export function draftFromConstraints(
  c: StringConstraints | undefined
): StringConstraintsDraft {
  return {
    minLength: c?.min_length != null ? String(c.min_length) : "",
    maxLength: c?.max_length != null ? String(c.max_length) : "",
    charactersLimited: (c?.allowed_characters ?? []).length > 0,
    allowedCharacters: c?.allowed_characters ?? [],
    pattern: c?.pattern ?? "",
  };
}

function parseLength(s: string): number | undefined {
  if (s.trim() === "") {
    return undefined;
  }
  const n = Number(s);
  return Number.isInteger(n) ? n : undefined;
}

// Only what is set: an empty field or "any character" writes no key.
export function constraintsFromDraft(
  d: StringConstraintsDraft
): StringConstraints {
  const out: StringConstraints = {};
  const min = parseLength(d.minLength);
  const max = parseLength(d.maxLength);
  if (min != null) out.min_length = min;
  if (max != null) out.max_length = max;
  const chars = ALLOWED_CHARACTERS.filter((c) =>
    d.allowedCharacters.includes(c)
  );
  if (d.charactersLimited && chars.length > 0) {
    out.allowed_characters = chars;
  }
  if (d.pattern !== "") out.pattern = d.pattern;
  return out;
}

export function hasConstraints(c: StringConstraints): boolean {
  return (
    c.min_length != null ||
    c.max_length != null ||
    (c.allowed_characters ?? []).length > 0 ||
    (c.pattern ?? "") !== ""
  );
}

export type SummaryPart =
  | { kind: "length-range"; min: number; max: number }
  | { kind: "length-min"; min: number }
  | { kind: "length-max"; max: number }
  | { kind: "characters"; characters: AllowedCharacter[] }
  | { kind: "pattern" };

// What a row says about its constraints, in order.
export function summaryParts(c: StringConstraints): SummaryPart[] {
  const out: SummaryPart[] = [];
  if (c.min_length != null && c.max_length != null) {
    out.push({ kind: "length-range", min: c.min_length, max: c.max_length });
  } else if (c.min_length != null) {
    out.push({ kind: "length-min", min: c.min_length });
  } else if (c.max_length != null) {
    out.push({ kind: "length-max", max: c.max_length });
  }
  if ((c.allowed_characters ?? []).length > 0) {
    out.push({ kind: "characters", characters: c.allowed_characters ?? [] });
  }
  if ((c.pattern ?? "") !== "") {
    out.push({ kind: "pattern" });
  }
  return out;
}

// The translation key of the message shown when the pattern does not match.
export function patternErrorKey(
  kind: "standard" | "custom",
  pointer: string
): string {
  return `${kind}-attribute-pattern-error-${pointer}`;
}

// standard_attributes.validation with each pointer's new constraints: in
// place of its entry, else appended; a pointer left with none has no entry.
export function applyValidationUpdates(
  list: StandardAttributesValidationConfig[],
  updates: { pointer: string; constraints: StringConstraints }[]
): StandardAttributesValidationConfig[] {
  const byPointer = new Map(updates.map((u) => [u.pointer, u.constraints]));
  const out: StandardAttributesValidationConfig[] = [];
  for (const entry of list) {
    const next = byPointer.get(entry.pointer);
    if (next == null) {
      out.push(entry);
    } else if (hasConstraints(next)) {
      out.push({ pointer: entry.pointer, ...next });
    }
    byPointer.delete(entry.pointer);
  }
  for (const [pointer, constraints] of byPointer) {
    if (hasConstraints(constraints)) {
      out.push({ pointer, ...constraints });
    }
  }
  return out;
}
