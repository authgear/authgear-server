import React from "react";
import {
  BaseEdge,
  Edge,
  EdgeLabelRenderer,
  EdgeProps,
  EdgeTypes,
} from "@xyflow/react";
import { curveLabelPoint, curvePath } from "./edgePath";
import { GAP_X } from "./layout";
import { InsertionPoint } from "./edit/pointer";
import { InsertButton } from "./edit/CanvasControls";
import { insertionKey, useStepDrag } from "./edit/StepDragContext";

// How far before the target the "after the choice" button sits.
const MERGE_BUTTON_OFFSET = 28;

// xyflow requires edge data to be a record.
export interface LaneEdgeData extends Record<string, unknown> {
  laneY: number | null;
  // Where a step added on this line goes, in edit mode.
  insertAt: InsertionPoint | null;
  hovered: boolean;
  // Where the options of a choice meet, just before the target: a step added
  // there runs after the choice, on every path.
  mergeInsertAt: InsertionPoint | null;
  // A step added on this line runs after the choice, for every option.
  afterChoice: boolean;
  // The line leaves an option: its "+" sits beside that option's output.
  fromOption: boolean;
  // Any line into this edge's target is hovered.
  targetHovered: boolean;
}
export type LaneEdge = Edge<LaneEdgeData>;

function LaneEdgeComponent({
  sourceX,
  sourceY,
  targetX,
  targetY,
  data,
}: EdgeProps<LaneEdge>): React.ReactElement {
  const geometry = {
    sourceX,
    sourceY,
    targetX,
    targetY,
    laneY: data?.laneY ?? null,
    gap: GAP_X,
  };
  const path = curvePath(geometry);
  const insertAt = data?.insertAt ?? null;
  const mergeInsertAt = data?.mergeInsertAt ?? null;
  // While a card is dragged, every "+" is a place to drop it.
  const drag = useStepDrag();
  const dropKey = drag?.target != null ? insertionKey(drag.target) : null;
  const { x: labelX, y: labelY } = curveLabelPoint(
    geometry,
    data?.fromOption ?? false
  );
  const mergeX = targetX - MERGE_BUTTON_OFFSET;
  const mergeY = targetY;
  return (
    <>
      <BaseEdge path={path} interactionWidth={24} />
      {insertAt != null ? (
        <EdgeLabelRenderer>
          <div
            data-insert-at={insertionKey(insertAt)}
            style={{
              position: "absolute",
              transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
            }}
          >
            <InsertButton
              at={insertAt}
              visible={(data?.hovered ?? false) || drag != null}
              highlighted={dropKey === insertionKey(insertAt)}
              afterChoice={data?.afterChoice ?? false}
            />
          </div>
        </EdgeLabelRenderer>
      ) : null}
      {mergeInsertAt != null ? (
        <EdgeLabelRenderer>
          <div
            data-insert-at={insertionKey(mergeInsertAt)}
            style={{
              position: "absolute",
              transform: `translate(-50%, -50%) translate(${mergeX}px, ${mergeY}px)`,
            }}
          >
            <InsertButton
              at={mergeInsertAt}
              visible={(data?.targetHovered ?? false) || drag != null}
              highlighted={dropKey === insertionKey(mergeInsertAt)}
              afterChoice={true}
            />
          </div>
        </EdgeLabelRenderer>
      ) : null}
    </>
  );
}

export const flowEdgeTypes: EdgeTypes = {
  lane: LaneEdgeComponent,
};
