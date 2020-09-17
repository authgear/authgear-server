/* global describe, it, expect */
import { Values } from "@oursky/react-messageformat";
import {
  violationSelector,
  makeMissingFieldSelector,
  errorFormatter,
} from "./validation";

const mockViolations = [
  {
    kind: "required" as const,
    missingField: ["key_id", "team_id"],
    location: "/identity/oauth/providers/0",
  },
];

const mockRenderToString = (messageId: string, values?: Values): string => {
  if (!(typeof values?.fieldName === "string")) {
    return messageId;
  }
  return `${messageId} ${values.fieldName}`;
};

describe("validation error handling tools", () => {
  it("violation selector with selector generated", () => {
    const violationMap = violationSelector(mockViolations, {
      key_id: makeMissingFieldSelector("/identity/oauth/providers", "key_id"),
      client_id: makeMissingFieldSelector(
        "/identity/oauth/providers",
        "client_id"
      ),
    });
    expect(violationMap["key_id"]).toEqual(
      expect.arrayContaining([
        {
          kind: "required" as const,
          missingField: ["key_id", "team_id"],
          location: "/identity/oauth/providers/0",
        },
      ])
    );
    expect(violationMap["client_id"].length).toEqual(0);
  });

  it("violation selector with custom selector", () => {
    const violationMap = violationSelector(mockViolations, {
      required: (violation) => violation.kind === "required",
      general: (violation) => violation.kind === "general",
    });
    expect(violationMap.required.length).toEqual(1);
    expect(violationMap.general.length).toEqual(0);
  });

  it("default error formatter required field missing", () => {
    // get default error message for supported kind of violation
    // by providing locale key of field name
    const errorMessageWithViolation = errorFormatter(
      "testing-field-locale-key",
      // violations of field
      mockViolations,
      mockRenderToString
    );
    expect(errorMessageWithViolation?.length).toBeGreaterThanOrEqual(1);

    const errorMessageWithoutViolation = errorFormatter(
      "testing field",
      // violations of field
      [],
      mockRenderToString
    );
    expect(errorMessageWithoutViolation).toBe(undefined);
  });
});
