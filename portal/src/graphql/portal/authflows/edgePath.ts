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

// One cubic S-curve from (x1, y1) to (x2, y2), leaving and arriving level.
function sCurve(x1: number, y1: number, x2: number, y2: number): string {
  const midX = (x1 + x2) / 2;
  return `C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`;
}

// An edge to the next column is a single S-curve. An edge that skips columns
// curves into its lane inside the source's gap, runs straight along the lane
// past the cards in between, and curves out inside the target's gap.
// The lane's y when the edge has room to run along it; null for an edge
// drawn as one curve.
function laneRunY({
  sourceX,
  targetX,
  laneY,
  gap,
}: CurvePathInput): number | null {
  return laneY != null && targetX - gap > sourceX + gap ? laneY : null;
}

// How far past an option's output its line's "+" sits.
export const OUTPUT_INSERT_OFFSET = 28;

// Where a line's "+" goes: beside the output it leaves, for a line that
// leaves an option, so the "+" is next to the option it adds to; otherwise
// on the drawn line, halfway along.
export function curveLabelPoint(
  input: CurvePathInput,
  besideSource = false
): {
  x: number;
  y: number;
} {
  if (besideSource) {
    return { x: input.sourceX + OUTPUT_INSERT_OFFSET, y: input.sourceY };
  }
  return {
    x: (input.sourceX + input.targetX) / 2,
    y: laneRunY(input) ?? (input.sourceY + input.targetY) / 2,
  };
}

export function curvePath(input: CurvePathInput): string {
  const { sourceX, sourceY, targetX, targetY, gap } = input;
  const laneY = laneRunY(input);
  const laneStartX = sourceX + gap;
  const laneEndX = targetX - gap;
  if (laneY == null) {
    return `M ${sourceX} ${sourceY} ${sCurve(
      sourceX,
      sourceY,
      targetX,
      targetY
    )}`;
  }
  return (
    `M ${sourceX} ${sourceY} ${sCurve(sourceX, sourceY, laneStartX, laneY)}` +
    ` L ${laneEndX} ${laneY} ${sCurve(laneEndX, laneY, targetX, targetY)}`
  );
}
