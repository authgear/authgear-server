import { describe, it, expect } from "@jest/globals";
import { isValidWebhookHookURI } from "./hookUri";

describe("isValidWebhookHookURI", () => {
  it("accepts empty value", () => {
    expect(isValidWebhookHookURI("")).toBe(true);
  });

  it("accepts valid http and https URLs", () => {
    expect(isValidWebhookHookURI("http://example.com")).toBe(true);
    expect(isValidWebhookHookURI("http://example.com/")).toBe(true);
    expect(isValidWebhookHookURI("http://example.com/a")).toBe(true);
    expect(isValidWebhookHookURI("http://example.com/a/")).toBe(true);
    expect(isValidWebhookHookURI("https://example.com/callback?a=b")).toBe(true);
  });

  it("rejects invalid URLs", () => {
    expect(isValidWebhookHookURI("https://")).toBe(false);
    expect(isValidWebhookHookURI("foobar")).toBe(false);
    expect(isValidWebhookHookURI("authgeardeno:///deno/a.ts")).toBe(false);
    expect(isValidWebhookHookURI("http://example.com/../secret")).toBe(false);
  });
});
