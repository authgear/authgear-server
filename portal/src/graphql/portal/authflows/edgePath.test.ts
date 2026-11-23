import { describe, it, expect } from "@jest/globals";
import { OUTPUT_INSERT_OFFSET, curveLabelPoint, curvePath } from "./edgePath";

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

describe("curveLabelPoint", () => {
  it("sits on the lane when the line runs along it", () => {
    expect(
      curveLabelPoint({
        sourceX: 0,
        sourceY: 0,
        targetX: 600,
        targetY: 0,
        laneY: 100,
        gap: 120,
      })
    ).toEqual({ x: 300, y: 100 });
  });

  it("sits on the curve when the line is too short for its lane", () => {
    expect(
      curveLabelPoint({
        sourceX: 0,
        sourceY: -80,
        targetX: 200,
        targetY: 40,
        laneY: 100,
        gap: 120,
      })
    ).toEqual({ x: 100, y: -20 });
  });
});

describe("curveLabelPoint beside the source", () => {
  it("sits just past the output a line leaves, level with it", () => {
    expect(
      curveLabelPoint(
        {
          sourceX: 0,
          sourceY: -80,
          targetX: 600,
          targetY: 40,
          laneY: 100,
          gap: 120,
        },
        true
      )
    ).toEqual({ x: OUTPUT_INSERT_OFFSET, y: -80 });
  });
});
