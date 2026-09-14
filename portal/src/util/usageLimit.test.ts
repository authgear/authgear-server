import { describe, it, expect } from "@jest/globals";
import { smallestBlockQuota } from "./usageLimit";

describe("smallestBlockQuota", () => {
  it("returns null when the plan carries no limit at all", () => {
    expect(smallestBlockQuota(undefined)).toEqual(null);
    expect(smallestBlockQuota([])).toEqual(null);
  });

  it("returns null when no entry blocks", () => {
    expect(
      smallestBlockQuota([
        { action: "alert", quota: 10 },
        { action: "alert", quota: 20 },
      ])
    ).toEqual(null);
  });

  it("ignores non-block entries even when they are smaller", () => {
    expect(
      smallestBlockQuota([
        { action: "alert", quota: 5 },
        { action: "block", quota: 50 },
      ])
    ).toEqual(50);
  });

  it("returns the smallest of several block entries", () => {
    expect(
      smallestBlockQuota([
        { action: "block", quota: 100 },
        { action: "block", quota: 25 },
        { action: "block", quota: 60 },
      ])
    ).toEqual(25);
  });

  // A block entry with no quota caps nothing, so it must not be mistaken for
  // a quota of 0 -- Math.min over an undefined would otherwise yield NaN.
  it("skips block entries with no quota", () => {
    expect(smallestBlockQuota([{ action: "block" }])).toEqual(null);
    expect(
      smallestBlockQuota([{ action: "block" }, { action: "block", quota: 30 }])
    ).toEqual(30);
  });

  it("treats an explicit quota of 0 as a real cap, not as absent", () => {
    expect(smallestBlockQuota([{ action: "block", quota: 0 }])).toEqual(0);
  });
});
