import { describe, it, expect } from "@jest/globals";
import {
  checkIsCustomDomain,
  getHostFromOrigin,
  getOriginFromDomain,
} from "./domain";
import { Domain } from "../graphql/portal/globalTypes.generated";

function domain(name: string, isCustom: boolean): Domain {
  return {
    id: name,
    createdAt: "2026-01-01T00:00:00Z",
    apexDomain: name,
    domain: name,
    cookieDomain: name,
    isCustom,
    isVerified: true,
    verificationDNSRecord: "",
  };
}

describe("Domain", () => {
  it("can get https url from domain", () => {
    expect(getOriginFromDomain("abc.com")).toEqual("https://abc.com");
  });

  it("can get host from url", () => {
    expect(getHostFromOrigin("https://abc.com")).toEqual("abc.com");
  });

  it("can get host with port from url", () => {
    expect(getHostFromOrigin("https://abc.com:3100")).toEqual("abc.com:3100");
  });

  describe("checkIsCustomDomain", () => {
    const domains = [
      domain("myapp.authgear.cloud", false),
      domain("example.com", true),
    ];

    it("matches a custom domain", () => {
      expect(checkIsCustomDomain(domains, "https://example.com")).toBe(true);
    });

    it("ignores the port of the public origin", () => {
      expect(checkIsCustomDomain(domains, "https://example.com:3000")).toBe(
        true
      );
    });

    it("does not match a default domain", () => {
      expect(checkIsCustomDomain(domains, "https://myapp.authgear.cloud")).toBe(
        false
      );
    });

    it("does not match an invalid public origin", () => {
      expect(checkIsCustomDomain(domains, "")).toBe(false);
    });
  });
});
