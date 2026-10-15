import React, { useContext } from "react";
import cn from "classnames";
import { Handle, Node, NodeProps, NodeTypes, Position } from "@xyflow/react";
import { Context } from "../../../intl";
import { Callout } from "../../../components/v2/Callout/Callout";
import { FlowStepItem } from "./tree";
import { FlowRef, NEXT_HANDLE_ID, optionHandleID } from "./graph";
import { flowSelectionSearch, resolveFlow } from "./model";
import Link from "../../../Link";
import {
  IfNeededMessageIDs,
  ifNeededMessageIDs,
  optionalMessageIDs,
  destinationMessageID,
  pagelessNoteMessageIDs,
  stepTitle,
} from "./labels";
import { Tooltip } from "../../../components/v2/Tooltip/Tooltip";
import {
  ContinueAsPage,
  StepSketch,
  continuePageOptionIndex,
} from "./sketches";
import { identifyGroupGeometry } from "./groupGeometry";
import { useFlowSketchContext } from "./FlowSketchContext";
import styles from "./FlowNodes.module.css";

// xyflow requires node data to be a record.
export interface FlowNodeData extends Record<string, unknown> {
  item: FlowStepItem | null;
  flowRef: FlowRef | null;
}
export type FlowNode = Node<FlowNodeData>;

function requireItem(data: FlowNodeData): FlowStepItem {
  if (data.item == null) {
    throw new Error("step node rendered without a step");
  }
  return data.item;
}

// The condition under which a step runs, for a step the runtime may skip:
// optional in config, or run only when needed. Null for a step that always runs.
function stepConditionMessageIDs(
  item: FlowStepItem
): IfNeededMessageIDs | null {
  if (item.optional) {
    return optionalMessageIDs(item.step);
  }
  return ifNeededMessageIDs(item.step.type);
}

// States when the step runs; the tooltip explains it in full.
function ConditionLine({
  item,
}: {
  item: FlowStepItem;
}): React.ReactElement | null {
  const { renderToString } = useContext(Context);
  const ids = stepConditionMessageIDs(item);
  if (ids == null) {
    return null;
  }
  return (
    <Tooltip content={renderToString(ids.note)}>
      <span className={styles.conditionLine}>
        <svg
          className={styles.conditionIcon}
          viewBox="0 0 16 16"
          aria-hidden={true}
        >
          <path d="M2 4h4c2 0 3 2 4 4s2 4 4 4" />
          <path d="M2 4h12" />
        </svg>
        {renderToString(ids.condition)}
      </span>
    </Tooltip>
  );
}

function PageNode({ data, selected }: NodeProps<FlowNode>): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { flowType, recoveryDestinations } = useFlowSketchContext();
  const item = requireItem(data);
  const isEmptyChoice = item.branches?.length === 0;
  const continueIndex = continuePageOptionIndex(item);
  if (continueIndex >= 0) {
    return (
      <IdentifyGroup
        item={item}
        continueIndex={continueIndex}
        selected={selected}
      />
    );
  }
  return (
    <div className={cn(styles.card, selected ? styles.selected : null)}>
      <Handle
        type="target"
        position={Position.Left}
        className={
          item.step.type === "verify_account_recovery_code" &&
          recoveryDestinations.length > 0
            ? styles.unusedHandle
            : undefined
        }
      />
      <ConditionLine item={item} />
      <div className={styles.cardHeader}>
        <span className={styles.cardTitle}>
          {stepTitle(item.step.type, renderToString)}
        </span>
      </div>
      {item.step.name != null ? (
        <span className={styles.stepName}>{item.step.name}</span>
      ) : null}
      {item.step.type === "fill_form" && item.step.form?.name != null ? (
        <span className={styles.stepName}>{item.step.form.name}</span>
      ) : null}
      <div className={styles.cardBody}>
        {isEmptyChoice ? (
          <Callout
            type="warning"
            size="1"
            text={renderToString(
              flowType === "account_recovery"
                ? "AuthFlowsScreen.empty-choice.account_recovery"
                : "AuthFlowsScreen.empty-choice.default"
            )}
          />
        ) : (
          <StepSketch item={item} hiddenIndex={-1} />
        )}
      </div>
      {item.branches == null ? (
        <Handle type="source" position={Position.Right} id={NEXT_HANDLE_ID} />
      ) : null}
    </div>
  );
}

const ARROW_MARKER_ID = "auth-flow-group-arrow";

// One identify step drawn as the two pages AuthUI shows, stacked: "Continue
// as", only when the user already has a session, above the login page.
function IdentifyGroup({
  item,
  continueIndex,
  selected,
}: {
  item: FlowStepItem;
  continueIndex: number;
  selected: boolean;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const g = identifyGroupGeometry();
  const continueRight = g.continueLeft + g.continueWidth;
  const actionCenterX = g.actionLeft + g.actionWidth / 2;
  return (
    <div
      className={cn(styles.group, selected ? styles.selected : null)}
      style={{
        width: g.width,
        paddingTop: g.loginTop,
        paddingLeft: g.loginLeft,
      }}
    >
      <div className={styles.groupHeader} style={{ height: g.headerHeight }}>
        <span className={styles.cardTitle}>
          {stepTitle(item.step.type, renderToString)}
        </span>
        {item.step.name != null ? (
          <span className={styles.stepName}>{item.step.name}</span>
        ) : null}
      </div>
      <svg className={styles.groupLines} aria-hidden={true}>
        <defs>
          <marker
            id={ARROW_MARKER_ID}
            viewBox="0 0 8 8"
            refX="7"
            refY="4"
            markerWidth="7"
            markerHeight="7"
            orient="auto"
          >
            <path className={styles.groupArrow} d="M0 0 L8 4 L0 8 z" />
          </marker>
        </defs>
        {/* The entry forks with two curves, bending around forkX. */}
        <path
          className={cn(styles.groupLine, styles.groupLineDashed)}
          d={`M0 ${g.entryY} C${g.forkX} ${g.entryY}, ${g.forkX} ${g.continueEntryY}, ${g.continueLeft} ${g.continueEntryY}`}
          markerEnd={`url(#${ARROW_MARKER_ID})`}
        />
        <path
          className={styles.groupLine}
          d={`M0 ${g.entryY} C${g.forkX} ${g.entryY}, ${g.forkX} ${g.loginEntryY}, ${g.loginLeft} ${g.loginEntryY}`}
          markerEnd={`url(#${ARROW_MARKER_ID})`}
        />
        <path
          className={styles.groupLine}
          d={`M${actionCenterX} ${g.continueTop + g.continueHeight} V${
            g.actionTop
          }`}
          markerEnd={`url(#${ARROW_MARKER_ID})`}
        />
        <path
          className={styles.groupLine}
          d={`M${actionCenterX} ${g.actionTop + g.actionHeight} V${g.loginTop}`}
          markerEnd={`url(#${ARROW_MARKER_ID})`}
        />
        <path
          className={styles.groupLine}
          d={`M${continueRight} ${g.continueButtonY} H${g.width}`}
        />
      </svg>
      <Handle
        type="target"
        position={Position.Left}
        style={{ top: g.entryY }}
      />
      <ContinueAsPage geometry={g} />
      <span
        className={styles.groupAction}
        style={{
          left: g.actionLeft,
          top: g.actionTop,
          width: g.actionWidth,
          height: g.actionHeight,
        }}
      >
        {renderToString("AuthFlowsScreen.sketch.use-another-account")}
      </span>
      {item.branches != null ? (
        <Handle
          type="source"
          position={Position.Right}
          id={optionHandleID(continueIndex)}
          style={{ top: g.continueButtonY }}
        />
      ) : null}
      <div
        className={styles.loginPage}
        style={{ width: g.loginWidth, marginBottom: g.padding }}
      >
        <span className={styles.pageCaption}>
          {renderToString("AuthFlowsScreen.sketch.otherwise-page")}
        </span>
        <StepSketch item={item} hiddenIndex={continueIndex} />
      </div>
    </div>
  );
}

const destinationTargetMessageIDs: Partial<Record<string, string>> = {
  email: "AuthFlowsScreen.destination.to.email",
  phone: "AuthFlowsScreen.destination.to.phone",
};

// select_destination with no page: what it sends, and that it goes to the
// email or phone the user just entered.
function DestinationLines({
  item,
}: {
  item: FlowStepItem;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const sends = (item.step.allowed_channels ?? []).map((c) => {
    const id = destinationMessageID(c.channel, c.otp_form);
    return id != null ? renderToString(id) : c.channel;
  });
  const target =
    item.identification != null
      ? destinationTargetMessageIDs[item.identification]
      : undefined;
  return (
    <Tooltip content={renderToString("AuthFlowsScreen.destination.pill-note")}>
      <span className={styles.pillNote}>
        {sends.join(" · ")}
        <br />
        {renderToString(target ?? "AuthFlowsScreen.destination.to.any")}
      </span>
    </Tooltip>
  );
}

function PagelessNode({
  data,
  selected,
}: NodeProps<FlowNode>): React.ReactElement {
  const { renderToString } = useContext(Context);
  const item = requireItem(data);
  const note = pagelessNoteMessageIDs(item.step);
  return (
    <div className={cn(styles.pill, selected ? styles.selected : null)}>
      <Handle type="target" position={Position.Left} />
      <ConditionLine item={item} />
      {stepTitle(item.step.type, renderToString)}
      {note != null ? (
        <Tooltip content={renderToString(note.note)}>
          <span className={styles.pillNote}>{renderToString(note.line)}</span>
        </Tooltip>
      ) : null}
      {item.step.type === "select_destination" ? (
        <DestinationLines item={item} />
      ) : null}
      <Handle type="source" position={Position.Right} id={NEXT_HANDLE_ID} />
    </div>
  );
}

function StartNode(): React.ReactElement {
  const { renderToString } = useContext(Context);
  return (
    <div className={styles.terminal}>
      {renderToString("AuthFlowsScreen.start")}
      <Handle type="source" position={Position.Right} id={NEXT_HANDLE_ID} />
    </div>
  );
}

function EndNode(): React.ReactElement {
  const { renderToString } = useContext(Context);
  return (
    <div className={styles.terminal}>
      <Handle type="target" position={Position.Left} />
      {renderToString("AuthFlowsScreen.end")}
    </div>
  );
}

const flowRefTitleMessageIDs: Record<FlowRef["type"], string> = {
  login: "AuthFlowsScreen.flow-node.login",
  signup: "AuthFlowsScreen.flow-node.signup",
};

// A flow this one hands off to. Opening it switches the canvas to that flow.
function FlowRefNode({ data }: NodeProps<FlowNode>): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { custom, generated } = useFlowSketchContext();
  const ref = data.flowRef;
  if (ref == null) {
    throw new Error("flow node rendered without a flow");
  }
  const exists = resolveFlow(custom, generated, ref.type, ref.name) != null;
  return (
    <div className={cn(styles.flowRef, exists ? null : styles.flowRefMissing)}>
      <Handle type="target" position={Position.Left} />
      <span className={styles.cardTitle}>
        {renderToString(flowRefTitleMessageIDs[ref.type])}
      </span>
      <span className={styles.stepName}>{ref.name}</span>
      {exists ? (
        <Link
          to={{ search: flowSelectionSearch(ref) }}
          className={cn(styles.flowRefOpen, "nodrag", "nopan")}
          onClick={(e) => e.stopPropagation()}
        >
          {renderToString("AuthFlowsScreen.flow-node.open")}
        </Link>
      ) : (
        <span className={styles.flowRefMissingText}>
          {renderToString("AuthFlowsScreen.flow-node.not-found")}
        </span>
      )}
    </div>
  );
}

// Keyed by GraphNodeKind. Unknown steps reuse PageNode, whose sketch falls back.
export const flowNodeTypes: NodeTypes = {
  start: StartNode,
  end: EndNode,
  flow: FlowRefNode,
  page: PageNode,
  unknown: PageNode,
  pageless: PagelessNode,
};
