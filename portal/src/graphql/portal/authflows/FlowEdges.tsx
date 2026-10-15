import React from "react";
import { BaseEdge, Edge, EdgeProps, EdgeTypes } from "@xyflow/react";
import { curvePath } from "./edgePath";
import { GAP_X } from "./layout";

// xyflow requires edge data to be a record.
export interface LaneEdgeData extends Record<string, unknown> {
  laneY: number | null;
}
export type LaneEdge = Edge<LaneEdgeData>;

function LaneEdgeComponent({
  sourceX,
  sourceY,
  targetX,
  targetY,
  data,
}: EdgeProps<LaneEdge>): React.ReactElement {
  const path = curvePath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    laneY: data?.laneY ?? null,
    gap: GAP_X,
  });
  return <BaseEdge path={path} />;
}

export const flowEdgeTypes: EdgeTypes = {
  lane: LaneEdgeComponent,
};
