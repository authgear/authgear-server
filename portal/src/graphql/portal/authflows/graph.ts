import { AuthenticationFlowOptionConfig } from "../../../types";
import { FlowStepItem, StepKind } from "./tree";

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

export type GraphNodeKind = "start" | "end" | "flow" | StepKind;

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
}

export function treeToGraph(tree: FlowStepItem[]): FlowGraph {
  const nodes: FlowGraphNode[] = [
    { id: START_NODE_ID, kind: "start", item: null },
  ];
  const edges: FlowGraphEdge[] = [];
  // Insertion-ordered, so flows appear in the order options first name them.
  const flowRefs = new Map<string, FlowRef>();

  const connect = (outlets: Outlet[], target: string) => {
    for (const o of outlets) {
      edges.push({
        id: `${o.source}:${o.sourceHandle}:${o.kind}->${target}`,
        ...o,
        target,
      });
    }
  };

  const walk = (sequence: FlowStepItem[], incoming: Outlet[]): Outlet[] => {
    let current = incoming;
    for (const item of sequence) {
      nodes.push({ id: item.id, kind: item.kind, item });
      connect(current, item.id);
      let outlets: Outlet[];
      if (item.branches == null) {
        outlets = [
          {
            source: item.id,
            sourceHandle: NEXT_HANDLE_ID,
            kind: "next",
            optionIndex: null,
            lane: null,
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
          };
          const refs = optionFlowRefs(b.option);
          if (refs.length === 0) {
            return walk(b.sequence, [start]);
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
            connect([{ ...start, sourceHandle: handle }], id);
          });
          return [];
        });
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
