export interface CurvePathInput {
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  // The y of the lane reserved for an edge that skips columns; null for an
  // edge to the next column.
  laneY: number | null;
  // The horizontal gap between columns. Curves bend only inside a gap.
  gap: number;
}

export interface CurvePath {
  path: string;
  labelX: number;
  labelY: number;
}

// One cubic S-curve from (x1, y1) to (x2, y2), leaving and arriving level.
function sCurve(x1: number, y1: number, x2: number, y2: number): string {
  const midX = (x1 + x2) / 2;
  return `C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`;
}

// An edge to the next column is a single S-curve. An edge that skips columns
// curves into its lane inside the source's gap, runs straight along the lane
// past the cards in between, and curves out inside the target's gap.
export function curvePath({
  sourceX,
  sourceY,
  targetX,
  targetY,
  laneY,
  gap,
}: CurvePathInput): CurvePath {
  const laneStartX = sourceX + gap;
  const laneEndX = targetX - gap;
  if (laneY == null || laneEndX <= laneStartX) {
    return {
      path: `M ${sourceX} ${sourceY} ${sCurve(
        sourceX,
        sourceY,
        targetX,
        targetY
      )}`,
      labelX: (sourceX + targetX) / 2,
      labelY: (sourceY + targetY) / 2,
    };
  }
  return {
    path:
      `M ${sourceX} ${sourceY} ${sCurve(sourceX, sourceY, laneStartX, laneY)}` +
      ` L ${laneEndX} ${laneY} ${sCurve(laneEndX, laneY, targetX, targetY)}`,
    labelX: (laneStartX + laneEndX) / 2,
    labelY: laneY,
  };
}
