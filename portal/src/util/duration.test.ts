import { describe, it, expect } from "@jest/globals";
import { parseDuration, formatDuration } from "./duration";

describe("parseDuration", () => {
  it("parse duration string as seconds", () => {
    expect(parseDuration("0")).toEqual(0);
    expect(parseDuration("-0")).toEqual(0);
    expect(parseDuration("3s")).toEqual(3);
    expect(parseDuration("2.0m")).toEqual(120);
    expect(parseDuration("0.03s")).toEqual(0.03);
    expect(parseDuration("7μs")).toEqual(0.000007);
    expect(parseDuration("-9h1s")).toEqual(-32401);
    expect(parseDuration("8m100s")).toEqual(580);
  });
  it("reject invalid duration string", () => {
    expect(() => parseDuration("1")).toThrow();
    expect(() => parseDuration("3d")).toThrow();
    expect(() => parseDuration("-1m-3s")).toThrow();
    expect(() => parseDuration("1e3s")).toThrow();
    expect(() => parseDuration(" 1s")).toThrow();
  });
});

describe("formatDuration", () => {
  it("format number as duration string", () => {
    expect(formatDuration(0)).toEqual("0s");
    expect(formatDuration(10)).toEqual("10s");
    expect(formatDuration(-2)).toEqual("-2s");
    expect(formatDuration(89.6)).toEqual("89.6s");
  });
});
