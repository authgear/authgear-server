import { describe, it, expect } from "@jest/globals";
import { formatSeconds } from "./formatDuration";

describe("formatSeconds", () => {
  it("should return null for negative numbers", () => {
    expect(formatSeconds("en-US", -1)).toBeNull();
    expect(formatSeconds("en-US", -100)).toBeNull();
    expect(formatSeconds("en-US", -0.5)).toBeNull();
  });

  it("should format zero seconds", () => {
    expect(formatSeconds("en-US", 0)).toBe("0 seconds");
  });

  it("should format seconds only", () => {
    expect(formatSeconds("en-US", 1)).toBe("1 second");
    expect(formatSeconds("en-US", 30)).toBe("30 seconds");
    expect(formatSeconds("en-US", 59)).toBe("59 seconds");
  });

  it("should format minutes and seconds", () => {
    expect(formatSeconds("en-US", 60)).toBe("1 minute");
    expect(formatSeconds("en-US", 61)).toBe("1 minute, 1 second");
    expect(formatSeconds("en-US", 90)).toBe("1 minute, 30 seconds");
    expect(formatSeconds("en-US", 120)).toBe("2 minutes");
    expect(formatSeconds("en-US", 3599)).toBe("59 minutes, 59 seconds");
  });

  it("should format hours, minutes and seconds", () => {
    expect(formatSeconds("en-US", 3600)).toBe("1 hour");
    expect(formatSeconds("en-US", 3661)).toBe("1 hour, 1 minute, 1 second");
    expect(formatSeconds("en-US", 7200)).toBe("2 hours");
    expect(formatSeconds("en-US", 7265)).toBe("2 hours, 1 minute, 5 seconds");
  });

  it("should format days, hours, minutes and seconds", () => {
    expect(formatSeconds("en-US", 86400)).toBe("1 day");
    expect(formatSeconds("en-US", 86461)).toBe("1 day, 1 minute, 1 second");
    expect(formatSeconds("en-US", 90061)).toBe(
      "1 day, 1 hour, 1 minute, 1 second"
    );
    expect(formatSeconds("en-US", 172800)).toBe("2 days");
  });

  it("should handle decimal seconds", () => {
    expect(formatSeconds("en-US", 1.5)).toBe("1.5 seconds");
    expect(formatSeconds("en-US", 0.5)).toBe("0.5 seconds");
  });

  it("should handle large numbers", () => {
    const result = formatSeconds("en-US", 999999);
    expect(result).toContain("days");
    expect(result).not.toBeNull();
  });

  it("should handle edge case of exactly 1 unit", () => {
    expect(formatSeconds("en-US", 1)).toBe("1 second");
    expect(formatSeconds("en-US", 60)).toBe("1 minute");
    expect(formatSeconds("en-US", 3600)).toBe("1 hour");
    expect(formatSeconds("en-US", 86400)).toBe("1 day");
  });
});
