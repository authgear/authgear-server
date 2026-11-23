import { NodeChange } from "@xyflow/react";

export interface SplitNodeChanges<N extends NodeChange> {
  // The canvas is read-only: only xyflow's measurements reach node state, so
  // no key (e.g. Backspace) can remove or move a node.
  measurements: N[];
  // undefined: selection untouched. null: cleared. string: newly selected node.
  selectedID: string | null | undefined;
}

// Selection is owned by the screen, so both clicks and keyboard Enter/Space,
// which xyflow reports only as select changes, open the step panel.
export function splitNodeChanges<N extends NodeChange>(
  changes: N[]
): SplitNodeChanges<N> {
  const measurements = changes.filter((c) => c.type === "dimensions");
  let selectedID: string | null | undefined;
  for (const c of changes) {
    if (c.type !== "select") {
      continue;
    }
    if (c.selected) {
      selectedID = c.id;
    } else if (selectedID === undefined) {
      selectedID = null;
    }
  }
  return { measurements, selectedID };
}

export interface MeasuredSize {
  width: number;
  height: number;
}

// Node sizes, kept across edits so the canvas relayouts in place instead of
// remounting. Returns prev itself when no size changed.
export function applySizes(
  prev: ReadonlyMap<string, MeasuredSize>,
  changes: NodeChange[]
): ReadonlyMap<string, MeasuredSize> {
  let next: Map<string, MeasuredSize> | null = null;
  for (const c of changes) {
    if (c.type !== "dimensions" || c.dimensions == null) {
      continue;
    }
    const old = (next ?? prev).get(c.id);
    if (
      old?.width === c.dimensions.width &&
      old.height === c.dimensions.height
    ) {
      continue;
    }
    next ??= new Map(prev);
    next.set(c.id, c.dimensions);
  }
  return next ?? prev;
}
