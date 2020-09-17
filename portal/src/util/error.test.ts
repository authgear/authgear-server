/* global describe, it, expect */
import { GraphQLError } from "graphql";
import { handleUpdateAppConfigError } from "./error";

const mockGraphQLError = new GraphQLError(
  "/authgear.yaml is invalid: invalid configuration:\n/identity/oauth/providers/0: required\n  map[actual:[alias client_id type] expected:[client_id key_id team_id type] missing:[key_id team_id]]",
  undefined,
  undefined,
  undefined,
  ["updateAppConfig"],
  undefined,
  {
    errorName: "Invalid",
    info: {
      causes: [
        {
          location: "/identity/oauth/providers/0",
          kind: "required",
          details: {
            actual: ["alias", "client_id", "type"],
            expected: ["client_id", "key_id", "team_id", "type"],
            missing: ["key_id", "team_id"],
          },
        },
      ],
    },
    reason: "ValidationFailed",
  }
);

describe("error parsing", () => {
  it("parse update app config missing field error", () => {
    const violations = handleUpdateAppConfigError(mockGraphQLError);
    const expectedViolations = [
      {
        kind: "required",
        missingField: ["key_id", "team_id"],
        location: "/identity/oauth/providers/0",
      },
    ];
    expect(violations).toEqual(expect.arrayContaining(expectedViolations));
  });
});
