import React, { useContext } from "react";
import { AuthenticationFlowOptionConfig } from "../../../../types";
import { InsertionPoint } from "./pointer";

export interface InsertableType {
  type: string;
  // Message ID saying why it cannot go here; null when it can.
  disabledReason: string | null;
}

// The first problem with a step: a save error, a check the server would
// reject, or a warning.
export interface StepProblem {
  tone: "error" | "warning";
  text: string;
}

// What the canvas can do in edit mode. Absent while only viewing.
export interface FlowEditorContextValue {
  insertableTypes: (at: InsertionPoint) => InsertableType[];
  // The option a step inserted at runs for; null outside an option.
  optionAt: (at: InsertionPoint) => AuthenticationFlowOptionConfig | null;
  onInsert: (at: InsertionPoint, type: string) => void;
  // Why the step cannot be removed, as rendered text; null when it can.
  removeBlockedReason: (stepID: string) => string | null;
  onRemove: (stepID: string) => void;
  // Moves a step, or copies it, to where a card was dropped.
  onMove: (stepID: string, to: InsertionPoint, copy: boolean) => void;
  // The same, for one option of a choice step.
  removeOptionBlockedReason: (stepID: string, index: number) => string | null;
  // Save errors and editor checks, by node.
  hasErrors: (stepID: string) => boolean;
  problem: (stepID: string) => StepProblem | null;
}

// Problems by step, in edit and read mode alike.
export interface StepProblemsContextValue {
  hasErrors: (stepID: string) => boolean;
  problem: (stepID: string) => StepProblem | null;
}

export const StepProblemsContext =
  React.createContext<StepProblemsContextValue | null>(null);

export const FlowEditorContext =
  React.createContext<FlowEditorContextValue | null>(null);

export function useFlowEditor(): FlowEditorContextValue | null {
  return useContext(FlowEditorContext);
}
