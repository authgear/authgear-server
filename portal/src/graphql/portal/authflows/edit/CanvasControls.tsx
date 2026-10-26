import React, { useContext, useState } from "react";
import { Button, DropdownMenu, Text } from "@radix-ui/themes";
import { NodeToolbar, Position } from "@xyflow/react";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import cn from "classnames";
import { Context } from "../../../../intl";
import { Tooltip } from "../../../../components/v2/Tooltip/Tooltip";
import { optionLabel, stepTitle } from "../labels";
import { useFlowSketchContext } from "../FlowSketchContext";
import { InsertionPoint } from "./pointer";
import {
  StepProblem,
  StepProblemsContext,
  useFlowEditor,
} from "./FlowEditorContext";
import styles from "./CanvasControls.module.css";

// Mounted only while the menu is open, so the step types are worked out
// against the current flow, not on every render of every line.
function InsertMenuItems({
  at,
  label,
}: {
  at: InsertionPoint;
  label: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const editor = useFlowEditor();
  return (
    <>
      <DropdownMenu.Label>{label}</DropdownMenu.Label>
      {(editor?.insertableTypes(at) ?? []).map(({ type, disabledReason }) => (
        <DropdownMenu.Item
          key={type}
          disabled={disabledReason != null}
          onSelect={() => editor?.onInsert(at, type)}
        >
          <span className={styles.insertItem}>
            {stepTitle(type, renderToString)}
            {disabledReason != null ? (
              <Text size="1" color="gray">
                {renderToString(disabledReason)}
              </Text>
            ) : null}
          </span>
        </DropdownMenu.Item>
      ))}
    </>
  );
}

// The "+" on a line: shown on hover and keyboard focus, it lists the step
// types that can go there.
export function InsertButton({
  at,
  visible,
  afterChoice = false,
  highlighted = false,
}: {
  at: InsertionPoint;
  visible: boolean;
  // A dragged card would land here.
  highlighted?: boolean;
  // Where a choice's options meet: the step goes on every option's path.
  afterChoice?: boolean;
}): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const { phoneOTPMode } = useFlowSketchContext();
  const editor = useFlowEditor();
  // Kept in state: the tooltip owns the button's data-state attribute.
  const [open, setOpen] = useState(false);
  if (editor == null) {
    return null;
  }
  // Says which options run the step, as a tooltip and the menu's heading.
  const option = afterChoice ? null : editor.optionAt(at);
  const label = afterChoice
    ? renderToString("AuthFlowsScreen.edit.insert.after-choice")
    : option != null
    ? renderToString("AuthFlowsScreen.edit.insert.for-option", {
        option: optionLabel(option, renderToString, phoneOTPMode),
      })
    : renderToString("AuthFlowsScreen.edit.insert");
  return (
    <DropdownMenu.Root open={open} onOpenChange={setOpen}>
      <Tooltip content={label}>
        <DropdownMenu.Trigger>
          <button
            type="button"
            className={`nodrag nopan ${styles.insertButton} ${
              visible || open ? styles.insertButtonVisible : ""
            } ${highlighted ? styles.insertButtonDropTarget : ""}`}
            aria-label={label}
          >
            +
          </button>
        </DropdownMenu.Trigger>
      </Tooltip>
      <DropdownMenu.Content size="1">
        <InsertMenuItems at={at} label={label} />
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  );
}

// Over the selected node: Remove, disabled with the reason when it would
// break the flow.
export function NodeEditToolbar({
  stepID,
  selected,
}: {
  stepID: string;
  selected: boolean;
}): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const editor = useFlowEditor();
  if (editor == null) {
    return null;
  }
  const blocked = editor.removeBlockedReason(stepID);
  return (
    <NodeToolbar isVisible={selected} position={Position.Top}>
      <Tooltip content={blocked} disabled={blocked == null}>
        <span>
          <Button
            type="button"
            size="1"
            variant="soft"
            color="red"
            disabled={blocked != null}
            onClick={() => editor.onRemove(stepID)}
          >
            {renderToString("AuthFlowsScreen.edit.remove")}
          </Button>
        </span>
      </Tooltip>
    </NodeToolbar>
  );
}

// Marks a step with a problem; selecting the step shows it in the panel.
export function ProblemBadge({
  stepID,
}: {
  stepID: string;
}): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const tone = useProblemTone(stepID);
  if (tone == null) {
    return null;
  }
  const label = renderToString(
    tone === "error"
      ? "AuthFlowsScreen.problem.error"
      : "AuthFlowsScreen.problem.warning"
  );
  return (
    <Tooltip content={label}>
      <span
        role="img"
        aria-label={label}
        className={cn(
          styles.problemBadge,
          tone === "error" ? styles.problemBadgeError : null
        )}
      >
        <ExclamationTriangleIcon width={11} height={11} />
      </span>
    </Tooltip>
  );
}

export function useProblemTone(stepID: string): StepProblem["tone"] | null {
  const problems = useContext(StepProblemsContext);
  if (problems == null) {
    return null;
  }
  if (problems.hasErrors(stepID)) {
    return "error";
  }
  return problems.problem(stepID)?.tone ?? null;
}
