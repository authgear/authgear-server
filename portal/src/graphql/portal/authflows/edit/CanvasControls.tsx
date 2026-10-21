import React, { useContext, useState } from "react";
import { Button, DropdownMenu, Text } from "@radix-ui/themes";
import { NodeToolbar, Position } from "@xyflow/react";
import { Context } from "../../../../intl";
import { Tooltip } from "../../../../components/v2/Tooltip/Tooltip";
import { optionLabel, stepTitle } from "../labels";
import { useFlowSketchContext } from "../FlowSketchContext";
import { InsertionPoint } from "./pointer";
import { useFlowEditor } from "./FlowEditorContext";
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

// The step's first problem, under its title.
export function WarningLine({
  stepID,
}: {
  stepID: string;
}): React.ReactElement | null {
  const editor = useFlowEditor();
  const problem = editor?.problem(stepID) ?? null;
  if (problem == null) {
    return null;
  }
  return (
    <span
      className={
        problem.tone === "error" ? styles.errorLine : styles.warningLine
      }
    >
      {problem.text}
    </span>
  );
}

export function useHasErrors(stepID: string): boolean {
  const editor = useFlowEditor();
  return editor?.hasErrors(stepID) ?? false;
}
