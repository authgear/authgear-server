import { AuthenticationFlowOptionConfig } from "../../../types";
import { FlowStepItem, StepKind, isDeviceTokenOption } from "./tree";
import { InsertionPoint, parentSteps } from "./edit/pointer";
import { AccountRecoveryDestination, selectDestinationChannels } from "./model";

export const START_NODE_ID = "start";
export const END_NODE_ID = "end";
export const NEXT_HANDLE_ID = "next";

export function optionHandleID(optionIndex: number): string {
  return `option-${optionIndex}`;
}

// An option that can hand off to two flows answers whether the account
// exists: AuthUI logs the user in when it does, and signs them up otherwise.
export type AccountAnswer = "has_account" | "no_account";

export function optionAnswerHandleID(
  optionIndex: number,
  answer: AccountAnswer
): string {
  return `${optionHandleID(optionIndex)}-${answer}`;
}

// The verify step of an account recovery flow has one section, and one
// input, per destination, so each select_destination lands on what it sends.
export function recoveryDestinationHandleID(
  d: AccountRecoveryDestination
): string {
  return `destination-${d.channel}-${d.otp_form}`;
}

export type GraphNodeKind = "start" | "end" | "flow" | "join" | StepKind;

// Where a boxed choice's options meet, on the right edge of its box.
export function joinNodeID(choiceID: string): string {
  return `${choiceID}:join`;
}

// Drawn as a box with one row per option when any option has steps of its
// own, so what sits inside a row runs for that option only.
export function isBoxedChoice(item: FlowStepItem): boolean {
  return item.branches?.some((b) => b.sequence.length > 0) ?? false;
}

// A flow another flow hands off to, as a signup_login option does.
export interface FlowRef {
  type: "login" | "signup";
  name: string;
}

export function flowRefNodeID(type: FlowRef["type"], name: string): string {
  return `flow:${type}:${name}`;
}

// The flows an option hands off to, login first.
export function optionFlowRefs(
  option: AuthenticationFlowOptionConfig
): FlowRef[] {
  const refs: FlowRef[] = [];
  if (option.login_flow != null) {
    refs.push({ type: "login", name: option.login_flow });
  }
  if (option.signup_flow != null) {
    refs.push({ type: "signup", name: option.signup_flow });
  }
  return refs;
}

// Options that can hand off to both flows, grouped by the two flows they
// name, in order of first appearance; each group lists option indexes. A
// group shares one has-account / no-account pair, from its first option.
export function handoffGroups(
  options: AuthenticationFlowOptionConfig[]
): number[][] {
  const groups = new Map<string, number[]>();
  options.forEach((option, i) => {
    if (option.login_flow == null || option.signup_flow == null) {
      return;
    }
    const key = JSON.stringify([option.login_flow, option.signup_flow]);
    const group = groups.get(key) ?? [];
    group.push(i);
    groups.set(key, group);
  });
  return [...groups.values()];
}

export interface FlowGraphNode {
  id: string;
  kind: GraphNodeKind;
  item: FlowStepItem | null;
  flowRef?: FlowRef;
}

export type FlowGraphEdgeKind = "next" | "option";

export interface FlowGraphEdge {
  id: string;
  source: string;
  sourceHandle: string;
  target: string;
  // null: the target's default input.
  targetHandle: string | null;
  // Where a step inserted on this line goes; null where none can.
  insertAt: InsertionPoint | null;
  // On one of the lines where a choice's options meet: a step inserted
  // there goes after the choice, on every option's path.
  mergeInsertAt: InsertionPoint | null;
  // A step inserted on this line goes after the choice, on every option's
  // path: the line leaving a boxed choice.
  afterChoice?: true;
  kind: FlowGraphEdgeKind;
  optionIndex: number | null;
  // The lane key layoutGraph reserves for an edge that must route around
  // other nodes; null for an edge that runs straight to its target.
  lane: string | null;
}

export function optionLaneKey(choiceID: string, optionIndex: number): string {
  return `${choiceID}/${optionHandleID(optionIndex)}`;
}

export interface FlowGraph {
  nodes: FlowGraphNode[];
  edges: FlowGraphEdge[];
  // The nodes a flow ends in, laid out together in the last column: End,
  // or the flows it hands off to.
  terminalIDs: string[];
}

// An outlet is an edge waiting for its target: whatever comes next.
interface Outlet {
  source: string;
  sourceHandle: string;
  kind: FlowGraphEdgeKind;
  optionIndex: number | null;
  lane: string | null;
  // What the select_destination this path last passed can send.
  destinations: AccountRecoveryDestination[] | null;
  insertAt: InsertionPoint | null;
  // Set on the outlets of a choice's options: the point after the choice.
  mergeAt: InsertionPoint | null;
  afterChoice?: true;
  // The option of the choice just walked that this path went through.
  fromOption?: number;
}

function optionSteps(stepID: string, optionIndex: number): string {
  return `${stepID}/one_of/${optionIndex}/steps`;
}

// Where a step inserted right after item goes. A step drawn with its only
// option inline continues into that option's steps, when it has any.
function insertionAfter(item: FlowStepItem): InsertionPoint | null {
  const options = item.step.one_of ?? [];
  const choices = options.filter((o) => !isDeviceTokenOption(o));
  if (item.branches == null && choices.length === 1) {
    // A hand-off ends the flow; nothing after it runs.
    if (optionFlowRefs(choices[0]).length > 0) {
      return null;
    }
    const j = options.indexOf(choices[0]);
    if ((choices[0].steps ?? []).length > 0) {
      return { steps: optionSteps(item.id, j), index: 0 };
    }
  }
  const { steps, index } = parentSteps(item.id);
  return { steps, index: index + 1 };
}

export function treeToGraph(tree: FlowStepItem[]): FlowGraph {
  const nodes: FlowGraphNode[] = [
    { id: START_NODE_ID, kind: "start", item: null },
  ];
  const edges: FlowGraphEdge[] = [];
  // Insertion-ordered, so flows appear in the order options first name them.
  const flowRefs = new Map<string, FlowRef>();

  const connect = (
    outlets: Outlet[],
    target: string,
    targetItem: FlowStepItem | null = null
  ) => {
    // Only the first line of each meeting carries its "after the choice".
    const marked = new Set<string>();
    for (const { destinations, mergeAt, fromOption: _, ...o } of outlets) {
      const id = `${o.source}:${o.sourceHandle}:${o.kind}->${target}`;
      const mergeKey = mergeAt != null ? JSON.stringify(mergeAt) : null;
      const mergeInsertAt =
        mergeKey != null && !marked.has(mergeKey) ? mergeAt : null;
      if (mergeKey != null) {
        marked.add(mergeKey);
      }
      if (
        targetItem?.step.type === "verify_account_recovery_code" &&
        destinations != null
      ) {
        destinations.forEach((d, k) => {
          const targetHandle = recoveryDestinationHandleID(d);
          edges.push({
            id: `${id}:${targetHandle}`,
            ...o,
            target,
            targetHandle,
            mergeInsertAt: k === 0 ? mergeInsertAt : null,
          });
        });
      } else {
        edges.push({ id, ...o, target, targetHandle: null, mergeInsertAt });
      }
    }
  };

  const walk = (sequence: FlowStepItem[], incoming: Outlet[]): Outlet[] => {
    let current = incoming;
    for (const item of sequence) {
      nodes.push({ id: item.id, kind: item.kind, item });
      connect(current, item.id, item);
      // Steps after a select_destination carry what it sends, until the
      // verify step they lead to.
      const incomingDestinations = new Set(current.map((o) => o.destinations));
      const destinations =
        item.step.type === "select_destination"
          ? selectDestinationChannels(item.step)
          : incomingDestinations.size === 1
          ? [...incomingDestinations][0]
          : null;
      let outlets: Outlet[];
      if (item.branches == null) {
        outlets = [
          {
            source: item.id,
            sourceHandle: NEXT_HANDLE_ID,
            kind: "next",
            optionIndex: null,
            lane: null,
            destinations,
            insertAt: insertionAfter(item),
            mergeAt: null,
          },
        ];
      } else {
        const followers = new Set(
          handoffGroups(item.step.one_of ?? []).flatMap((g) => g.slice(1))
        );
        outlets = item.branches.flatMap((b) => {
          const start: Outlet = {
            source: item.id,
            sourceHandle: optionHandleID(b.optionIndex),
            kind: "option",
            optionIndex: b.optionIndex,
            lane: optionLaneKey(item.id, b.optionIndex),
            destinations,
            insertAt: { steps: optionSteps(item.id, b.optionIndex), index: 0 },
            mergeAt: null,
          };
          const refs = optionFlowRefs(b.option);
          if (refs.length === 0) {
            return walk(b.sequence, [start]).map((o) => ({
              ...o,
              fromOption: b.optionIndex,
            }));
          }
          // A grouped option's lines leave from its group's first option.
          if (followers.has(b.optionIndex)) {
            return [];
          }
          // A hand-off ends this flow: the option leads to the named flows.
          // Login comes first in optionFlowRefs, so it is the has-account
          // answer when the option can lead to both.
          const answers: AccountAnswer[] = ["has_account", "no_account"];
          refs.forEach((ref, k) => {
            const id = flowRefNodeID(ref.type, ref.name);
            flowRefs.set(id, ref);
            const handle =
              refs.length > 1
                ? optionAnswerHandleID(b.optionIndex, answers[k])
                : start.sourceHandle;
            // signup_login options cannot have steps.
            connect([{ ...start, sourceHandle: handle, insertAt: null }], id);
          });
          return [];
        });
      }
      if (item.branches != null) {
        const { steps, index } = parentSteps(item.id);
        const afterChoice = { steps, index: index + 1 };
        // Lines into account recovery's verify step land on what each
        // option sends, so they cannot meet first.
        const joins =
          isBoxedChoice(item) &&
          outlets.length > 0 &&
          outlets.every((o) => o.destinations == null);
        if (joins) {
          const join = joinNodeID(item.id);
          nodes.push({ id: join, kind: "join", item: null });
          // Each option's line keeps to its row until the exit.
          connect(
            outlets.map((o) => ({
              ...o,
              lane:
                o.fromOption != null
                  ? optionLaneKey(item.id, o.fromOption)
                  : o.lane,
            })),
            join
          );
          outlets = [
            {
              source: join,
              sourceHandle: NEXT_HANDLE_ID,
              kind: "next",
              optionIndex: null,
              lane: null,
              destinations: null,
              insertAt: afterChoice,
              mergeAt: null,
              afterChoice: true,
            },
          ];
        } else {
          outlets = outlets.map((o) => ({ ...o, mergeAt: afterChoice }));
        }
      }
      current = outlets;
    }
    return current;
  };

  const last = walk(tree, [
    {
      source: START_NODE_ID,
      sourceHandle: NEXT_HANDLE_ID,
      kind: "next",
      optionIndex: null,
      lane: null,
      destinations: null,
      insertAt: { steps: "/steps", index: 0 },
      mergeAt: null,
    },
  ]);
  const terminalIDs: string[] = [];
  for (const [id, flowRef] of flowRefs) {
    nodes.push({ id, kind: "flow", item: null, flowRef });
    terminalIDs.push(id);
  }
  // A flow whose every path hands off has nothing reaching End.
  if (last.length > 0 || flowRefs.size === 0) {
    nodes.push({ id: END_NODE_ID, kind: "end", item: null });
    connect(last, END_NODE_ID);
    terminalIDs.push(END_NODE_ID);
  }
  return { nodes, edges, terminalIDs };
}
