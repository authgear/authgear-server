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
