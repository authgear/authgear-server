/* global describe, it, expect */
import { ApolloError } from "@apollo/client";
import { parseRawError, parseAPIErrors } from "./parse";

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
