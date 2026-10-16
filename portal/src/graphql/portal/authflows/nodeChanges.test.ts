import { describe, it, expect } from "@jest/globals";
import { NodeChange } from "@xyflow/react";
import { applySizes, splitNodeChanges } from "./nodeChanges";

describe("splitNodeChanges", () => {
  it("keeps only measurements, so the canvas cannot lose nodes", () => {
    const changes: NodeChange[] = [
      { type: "dimensions", id: "a", dimensions: { width: 1, height: 2 } },
      { type: "remove", id: "b" },
      { type: "position", id: "c", position: { x: 1, y: 1 } },
    ];
    expect(splitNodeChanges(changes).measurements).toEqual([changes[0]]);
  });

  it("reports the newly selected node, whether by click or keyboard", () => {
    const changes: NodeChange[] = [
      { type: "select", id: "a", selected: false },
      { type: "select", id: "b", selected: true },
    ];
    expect(splitNodeChanges(changes).selectedID).toEqual("b");
  });

  it("reports a cleared selection", () => {
    const changes: NodeChange[] = [
      { type: "select", id: "a", selected: false },
    ];
    expect(splitNodeChanges(changes).selectedID).toBeNull();
  });

  it("leaves the selection alone when nothing was selected or cleared", () => {
    const changes: NodeChange[] = [
      { type: "dimensions", id: "a", dimensions: { width: 1, height: 2 } },
    ];
    expect(splitNodeChanges(changes).selectedID).toBeUndefined();
  });
});

describe("applySizes", () => {
  it("records measured sizes and keeps the map when nothing changed", () => {
    const prev = new Map([["a", { width: 1, height: 2 }]]);
    const same = applySizes(prev, [
      { type: "dimensions", id: "a", dimensions: { width: 1, height: 2 } },
    ]);
    expect(same).toBe(prev);
    const next = applySizes(prev, [
      { type: "dimensions", id: "b", dimensions: { width: 3, height: 4 } },
      { type: "dimensions", id: "c" },
    ]);
    expect(next).toEqual(
      new Map([
        ["a", { width: 1, height: 2 }],
        ["b", { width: 3, height: 4 }],
      ])
    );
  });
});
