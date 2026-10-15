import { describe, it, expect } from "@jest/globals";
import { curvePath } from "./edgePath";

describe("curvePath", () => {
  it("draws one S-curve between neighbouring columns", () => {
    expect(
      curvePath({
        sourceX: 0,
        sourceY: 0,
        targetX: 120,
        targetY: 40,
        laneY: null,
        gap: 120,
      })
    ).toEqual("M 0 0 C 60 0, 60 40, 120 40");
  });

  it("curves into its lane in the source's gap and out in the target's", () => {
    expect(
      curvePath({
        sourceX: 0,
        sourceY: 0,
        targetX: 600,
        targetY: 0,
        laneY: -100,
        gap: 120,
      })
    ).toEqual(
      "M 0 0 C 60 0, 60 -100, 120 -100 L 480 -100 C 540 -100, 540 0, 600 0"
    );
  });

  it("ignores the lane when there is no room for it", () => {
    expect(
      curvePath({
        sourceX: 0,
        sourceY: 0,
        targetX: 120,
        targetY: 0,
        laneY: -100,
        gap: 120,
      })
    ).toEqual("M 0 0 C 60 0, 60 0, 120 0");
  });
});
