/* global describe, it, expect */
import { ApolloError } from "@apollo/client";
import {
  parseRawError,
  parseAPIErrors,
  makeInvariantViolatedErrorParseRule,
} from "./parse";
import { APIError } from "./error";

function apolloNetworkError(statusCode: number): ApolloError {
  const networkError = new Error("network error") as Error & {
    statusCode: number;
  };
  networkError.statusCode = statusCode;
  return new ApolloError({ networkError });
}

function topMessageIDs(error: unknown): (string | undefined)[] {
  const { topErrors } = parseAPIErrors(parseRawError(error), [], []);
  return topErrors.map((e) => e.messageID);
}

describe("parseRawError", () => {
  it("maps a 429 to RateLimited rather than a network failure", () => {
    // A rate limited request is rejected before it reaches the GraphQL
    // executor, so it arrives as an HTTP error rather than in the GraphQL
    // errors array.
    expect(parseRawError(apolloNetworkError(429))).toEqual([
      { reason: "RateLimited", errorName: "TooManyRequest" },
    ]);
  });

  it("maps a 413 to RequestEntityTooLarge", () => {
    expect(parseRawError(apolloNetworkError(413))).toEqual([
      { reason: "RequestEntityTooLarge", errorName: "RequestEntityTooLarge" },
    ]);
  });

  it("maps any other network error to NetworkFailed", () => {
    expect(parseRawError(apolloNetworkError(500))).toEqual([
      { reason: "NetworkFailed", errorName: "NetworkFailed" },
    ]);
  });
});

describe("parseAPIErrors", () => {
  it("shows the retry message for a rate limited request", () => {
    expect(topMessageIDs(apolloNetworkError(429))).toEqual([
      "errors.rate-limited",
    ]);
  });

  it("shows the retry message for a TooManyRequest API error", () => {
    expect(
      topMessageIDs({ reason: "TooManyRequest", errorName: "TooManyRequest" })
    ).toEqual(["errors.rate-limited"]);
  });

  it("still shows the network message for other network errors", () => {
    expect(topMessageIDs(apolloNetworkError(500))).toEqual(["errors.network"]);
  });
});

function topErrorsOf(error: APIError) {
  return parseAPIErrors([error], [], []).topErrors;
}

describe("parseAPIErrors without a screen rule", () => {
  it("names the login ID type for a duplicated identity", () => {
    const [e] = topErrorsOf({
      errorName: "Invalid",
      reason: "InvariantViolated",
      info: {
        cause: { kind: "DuplicatedIdentity" },
        IdentityTypeIncoming: "login_id",
        LoginIDTypeIncoming: "username",
      },
    });
    expect(e.messageID).toEqual("errors.invariant.duplicated-identity");
    expect(e.arguments).toMatchObject({ type: "username" });
  });

  it("uses the kind message for a reason it does not know", () => {
    const [e] = topErrorsOf({
      errorName: "NotFound",
      reason: "RoleNotFound",
    } as unknown as APIError);
    expect(e.messageID).toEqual("errors.not-found");
    expect(e.arguments).toMatchObject({ code: "RoleNotFound" });
  });

  it("shows the server message for a generic Invalid error", () => {
    const [e] = topErrorsOf({
      errorName: "Invalid",
      reason: "Invalid",
      message: "custom domain is not supported",
    } as unknown as APIError);
    expect(e).toEqual({ message: "custom domain is not supported" });
  });

  it("falls back to the generic message with the reason as the code", () => {
    const [e] = topErrorsOf({
      errorName: "Invalid",
      reason: "UserIsAnonymized",
      message: "user is anonymized",
    } as unknown as APIError);
    expect(e.messageID).toEqual("errors.unknown");
    expect(e.arguments).toMatchObject({ code: "UserIsAnonymized" });
  });

  it("includes the cause kind in the code for an unmapped invariant", () => {
    const [e] = topErrorsOf({
      errorName: "Invalid",
      reason: "InvariantViolated",
      info: { cause: { kind: "SomethingNew" } },
    } as unknown as APIError);
    expect(e.messageID).toEqual("errors.unknown");
    expect(e.arguments).toMatchObject({
      code: "InvariantViolated: SomethingNew",
    });
  });

  it("shows the generic message without a code for a non-API error", () => {
    expect(topMessageIDs(new Error("boom"))).toEqual(["errors.unexpected"]);
  });
});

describe("parseAPIErrors with a screen rule", () => {
  it("lets the screen rule override the shared message", () => {
    const { topErrors } = parseAPIErrors(
      [
        {
          errorName: "Invalid",
          reason: "InvariantViolated",
          info: { cause: { kind: "RemoveLastIdentity" } },
        },
      ],
      [],
      [makeInvariantViolatedErrorParseRule("RemoveLastIdentity", "custom")]
    );
    expect(topErrors.map((e) => e.messageID)).toEqual(["custom"]);
  });
});

describe("parseAPIErrors for specific reasons", () => {
  it("reports an email usage limit instead of asking to retry", () => {
    const [e] = topErrorsOf({
      errorName: "TooManyRequest",
      reason: "UsageLimitExceeded",
      info: { usage_name: "email", period: "month" },
    });
    expect(e.messageID).toEqual("errors.usage-limit-exceeded");
    expect(e.arguments).toMatchObject({ usageName: "email", period: "month" });
  });

  it("shows the script's error for a failed Deno run", () => {
    const [e] = topErrorsOf({
      errorName: "BadRequest",
      reason: "DenoRunError",
      message: "TypeError: x is undefined",
    });
    expect(e.messageID).toEqual("errors.deno-hook.run");
    expect(e.arguments).toMatchObject({
      message: "TypeError: x is undefined",
    });
  });

  it("names a duplicated scope", () => {
    expect(
      topErrorsOf({ errorName: "BadRequest", reason: "ScopeDuplicate" }).map(
        (e) => e.messageID
      )
    ).toEqual(["errors.scope-duplicate"]);
  });
});
