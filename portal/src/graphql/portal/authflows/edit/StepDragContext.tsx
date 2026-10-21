import React, { useContext } from "react";
import { InsertionPoint } from "./pointer";

// The card being dragged in Lanes, and the "+" it would land on if dropped.
export interface StepDrag {
  stepID: string;
  target: InsertionPoint | null;
}

export const StepDragContext = React.createContext<StepDrag | null>(null);

export function useStepDrag(): StepDrag | null {
  return useContext(StepDragContext);
}

// Insertion points are compared by value: the edges render their own copies.
export function insertionKey(at: InsertionPoint): string {
  return `${at.steps}#${at.index}`;
}
