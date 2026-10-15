import { FlowStepItem } from "./tree";
import { START_NODE_ID, optionLaneKey } from "./graph";

export interface Size {
  width: number;
  height: number;
}

export interface Position {
  x: number;
  y: number;
}

export const GAP_X = 120;
export const GAP_Y = 40;
// Height reserved for the edge of an empty branch beside non-empty ones, which
// runs past the cards of its siblings.
export const LANE_HEIGHT = 32;

export interface FlowLayout {
  // Top-left corner of every node.
  positions: Map<string, Position>;
  // The y of every reserved lane, keyed by FlowGraphEdge.lane.
  lanes: Map<string, number>;
}

interface ItemBox {
  box: Size;
  // Total height of the branches stacked under a choice.
  stackHeight: number;
  branchBoxes: Size[];
}

// Lays the tree out left to right. A choice stacks its branches top to bottom
// in config order, and the step after it starts right of the widest branch.
export function layoutGraph(
  tree: FlowStepItem[],
  sizes: ReadonlyMap<string, Size>,
  // Placed together after the tree, stacked top to bottom: End, or the
  // flows the tree hands off to.
  terminalIDs: string[]
): FlowLayout {
  const sizeOf = (id: string): Size => {
    const size = sizes.get(id);
    if (size == null) {
      throw new Error(`layoutGraph: no size for node ${id}`);
    }
    return size;
  };

  const itemBoxes = new Map<FlowStepItem, ItemBox>();
  const itemBoxOf = (item: FlowStepItem): ItemBox => {
    const b = itemBoxes.get(item);
    if (b == null) {
      throw new Error(`layoutGraph: ${item.id} was not measured`);
    }
    return b;
  };

  // Pass 1: measure every item bottom-up.
  const measureSequence = (sequence: FlowStepItem[]): Size => {
    let box: Size = { width: 0, height: 0 };
    sequence.forEach((item, i) => {
      const node = sizeOf(item.id);
      const contentBoxes = (item.branches ?? []).map((b) =>
        measureSequence(b.sequence)
      );
      const branchWidth = Math.max(0, ...contentBoxes.map((b) => b.width));
      // When every branch is empty, its edges go straight to the next column
      // and cross nothing, so they need no lane.
      const branchBoxes = contentBoxes.map((b) =>
        b.width === 0 && branchWidth > 0 ? { width: 0, height: LANE_HEIGHT } : b
      );
      const stackHeight = branchBoxes.reduce(
        (sum, b, j) => sum + (j > 0 ? GAP_Y : 0) + b.height,
        0
      );
      const itemBox: Size = {
        width: node.width + (branchWidth > 0 ? GAP_X + branchWidth : 0),
        height: Math.max(node.height, stackHeight),
      };
      itemBoxes.set(item, {
        box: itemBox,
        stackHeight,
        branchBoxes,
      });
      box = {
        width: box.width + (i > 0 ? GAP_X : 0) + itemBox.width,
        height: Math.max(box.height, itemBox.height),
      };
    });
    return box;
  };

  const positions = new Map<string, Position>();
  const lanes = new Map<string, number>();
  const placeNode = (id: string, x: number, centerY: number) => {
    positions.set(id, { x, y: centerY - sizeOf(id).height / 2 });
  };

  // Pass 2: place items and lanes. Returns the x just past the last item.
  const placeSequence = (
    sequence: FlowStepItem[],
    x: number,
    centerY: number
  ): number => {
    let cursor = x;
    sequence.forEach((item, i) => {
      if (i > 0) {
        cursor += GAP_X;
      }
      placeNode(item.id, cursor, centerY);
      const { box, stackHeight, branchBoxes } = itemBoxOf(item);
      const branchX = cursor + sizeOf(item.id).width + GAP_X;
      let top = centerY - stackHeight / 2;
      (item.branches ?? []).forEach((b, j) => {
        const height = branchBoxes[j].height;
        if (b.sequence.length === 0) {
          if (height > 0) {
            lanes.set(optionLaneKey(item.id, b.optionIndex), top + height / 2);
          }
        } else {
          placeSequence(b.sequence, branchX, top + height / 2);
        }
        top += height + GAP_Y;
      });
      cursor += box.width;
    });
    return cursor;
  };

  placeNode(START_NODE_ID, 0, 0);
  measureSequence(tree);
  const firstX = sizeOf(START_NODE_ID).width + GAP_X;
  const afterTree = placeSequence(tree, firstX, 0);
  const terminalX = tree.length > 0 ? afterTree + GAP_X : firstX;
  const terminalsHeight = terminalIDs.reduce(
    (sum, id, i) => sum + (i > 0 ? GAP_Y : 0) + sizeOf(id).height,
    0
  );
  let terminalTop = -terminalsHeight / 2;
  for (const id of terminalIDs) {
    const height = sizeOf(id).height;
    placeNode(id, terminalX, terminalTop + height / 2);
    terminalTop += height + GAP_Y;
  }
  return { positions, lanes };
}
