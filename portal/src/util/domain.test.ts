import { describe, it, expect } from "@jest/globals";
import { getHostFromOrigin, getOriginFromDomain } from "./domain";

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
});
