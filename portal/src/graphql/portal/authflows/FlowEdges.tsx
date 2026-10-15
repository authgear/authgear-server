import React, { useContext } from "react";
import { BaseEdge, Edge, EdgeProps, EdgeTypes } from "@xyflow/react";
import { Context } from "../../../intl";
import { curvePath } from "./edgePath";
import { GAP_X } from "./layout";

const SKIP_EDGE_DASH = "6 4";

// xyflow requires edge data to be a record.
export interface LaneEdgeData extends Record<string, unknown> {
  laneY: number | null;
  skip: boolean;
}
export type LaneEdge = Edge<LaneEdgeData>;

function LaneEdgeComponent({
  sourceX,
  sourceY,
  targetX,
  targetY,
  data,
}: EdgeProps<LaneEdge>): React.ReactElement {
  const { renderToString } = useContext(Context);
  const skip = data?.skip === true;
  const { path, labelX, labelY } = curvePath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    laneY: data?.laneY ?? null,
    gap: GAP_X,
  });
  return (
    <BaseEdge
      path={path}
      labelX={labelX}
      labelY={labelY}
      label={skip ? renderToString("AuthFlowsScreen.edge.skip") : undefined}
      style={skip ? { strokeDasharray: SKIP_EDGE_DASH } : undefined}
    />
  );
}

export const flowEdgeTypes: EdgeTypes = {
  lane: LaneEdgeComponent,
};
