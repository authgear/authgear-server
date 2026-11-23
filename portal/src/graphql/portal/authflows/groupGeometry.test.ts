import { describe, it, expect } from "@jest/globals";
import { identifyGroupGeometry } from "./groupGeometry";

describe("identifyGroupGeometry", () => {
  const g = identifyGroupGeometry();
  const continueBottom = g.continueTop + g.continueHeight;

  it("stacks the login page below the continue page", () => {
    expect(g.loginTop).toBeGreaterThan(continueBottom);
    expect(g.loginLeft).toEqual(g.continueLeft);
    expect(g.loginWidth).toEqual(g.continueWidth);
  });

  it("enters the step between the two pages", () => {
    expect(g.entryY).toBeGreaterThan(continueBottom);
    expect(g.entryY).toBeLessThan(g.loginTop);
    expect(g.forkX).toBeLessThan(g.continueLeft);
  });

  it("puts the use-another-account node between the two pages", () => {
    expect(g.actionTop).toBeGreaterThan(continueBottom);
    expect(g.actionTop + g.actionHeight).toBeLessThan(g.loginTop);
    expect(g.actionLeft).toBeGreaterThanOrEqual(g.continueLeft);
    expect(g.actionLeft + g.actionWidth).toBeLessThanOrEqual(
      g.continueLeft + g.continueWidth
    );
  });

  it("enters the step beside the use-another-account node, not through it", () => {
    expect(g.forkX).toBeLessThan(g.actionLeft);
  });

  it("keeps the continue line inside the continue page", () => {
    expect(g.continueButtonY).toBeGreaterThan(g.continueTop);
    expect(g.continueButtonY).toBeLessThan(continueBottom);
  });

  it("leaves room on the right for the handles", () => {
    expect(g.width).toBeGreaterThan(g.loginLeft + g.loginWidth);
  });
});
