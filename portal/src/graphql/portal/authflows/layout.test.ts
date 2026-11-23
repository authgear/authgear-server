import { describe, it, expect } from "@jest/globals";
import { flowToTree } from "./tree";
import {
  END_NODE_ID,
  FlowGraph,
  flowRefNodeID,
  joinNodeID,
  optionLaneKey,
  treeToGraph,
} from "./graph";
import {
  BOX_EXIT,
  BOX_PAD,
  GAP_X,
  GAP_Y,
  Position,
  Size,
  layoutGraph,
} from "./layout";
import {
  generatedEmailOAuthPasskey2FA,
  generatedUsernameOnly,
} from "./__fixtures__/flows";
import { AuthenticationFlowConfig } from "../../../types";

const W = 200;
const H = 100;
// One column: a node plus the gap after it.
const COLUMN = W + GAP_X;

function only<T>(xs: T[] | undefined): T {
  if (xs?.length !== 1) {
    throw new Error("expected exactly one element");
  }
  return xs[0];
}

// The dot where a box's rows meet.
const JOIN = 10;

function sizeOfKind(kind: string): Size {
  return kind === "join"
    ? { width: JOIN, height: JOIN }
    : { width: W, height: H };
}

function uniformSizes(graph: FlowGraph): Map<string, Size> {
  return new Map(graph.nodes.map((n) => [n.id, sizeOfKind(n.kind)]));
}

function rectOf(
  graph: FlowGraph,
  positions: Map<string, Position>,
  id: string
) {
  const kind = graph.nodes.find((n) => n.id === id)?.kind ?? "";
  return { id, ...at(positions, id), ...sizeOfKind(kind) };
}

function layout(flow: AuthenticationFlowConfig) {
  const tree = flowToTree(flow);
  const graph = treeToGraph(tree);
  const { positions, lanes } = layoutGraph(
    tree,
    uniformSizes(graph),
    graph.terminalIDs
  );
  return { graph, positions, lanes };
}

function at(positions: Map<string, Position>, id: string): Position {
  const p = positions.get(id);
  if (p == null) {
    throw new Error(`no position for ${id}`);
  }
  return p;
}

function expectNoOverlap(graph: FlowGraph, positions: Map<string, Position>) {
  const rects = graph.nodes.map((n) => rectOf(graph, positions, n.id));
  for (let i = 0; i < rects.length; i++) {
    for (let j = i + 1; j < rects.length; j++) {
      const a = rects[i];
      const b = rects[j];
      const overlaps =
        a.x < b.x + b.width &&
        b.x < a.x + a.width &&
        a.y < b.y + b.height &&
        b.y < a.y + a.height;
      if (overlaps) {
        throw new Error(`${a.id} overlaps ${b.id}`);
      }
    }
  }
}

// Checks the long horizontal run of every edge (see curvePath): it must not
// cross a card other than its own source and target.
function expectEdgesClearOfCards({
  graph,
  positions,
  lanes,
}: ReturnType<typeof layout>) {
  for (const e of graph.edges) {
    const s = rectOf(graph, positions, e.source);
    const t = at(positions, e.target);
    const from = s.x + s.width + GAP_X / 2;
    const to = t.x - GAP_X / 2;
    if (to - from <= 0) {
      continue;
    }
    const laneY = e.lane != null ? lanes.get(e.lane) : undefined;
    if (laneY == null && e.kind !== "next") {
      throw new Error(`${e.id} skips columns without a lane`);
    }
    const y = laneY ?? s.y + s.height / 2;
    for (const n of graph.nodes) {
      if (n.id === e.source || n.id === e.target) {
        continue;
      }
      const r = rectOf(graph, positions, n.id);
      if (r.x < to && r.x + r.width > from && r.y < y && y < r.y + r.height) {
        throw new Error(`${e.id} runs under ${n.id}`);
      }
    }
  }
}

const optionalInBranch: AuthenticationFlowConfig = {
  name: "default",
  steps: [
    {
      type: "authenticate",
      one_of: [
        {
          authentication: "primary_password",
          steps: [
            {
              type: "authenticate",
              optional: true,
              one_of: [
                { authentication: "secondary_totp" },
                { authentication: "recovery_code" },
              ],
            },
            { type: "change_password" },
          ],
        },
        { authentication: "primary_passkey" },
      ],
    },
    { type: "check_account_status" },
  ],
};

const branchy: AuthenticationFlowConfig = {
  name: "default",
  steps: [
    {
      type: "identify",
      one_of: [
        {
          identification: "email",
          steps: [
            { type: "verify" },
            {
              type: "create_authenticator",
              one_of: [{ authentication: "primary_password" }],
            },
          ],
        },
        { identification: "phone", steps: [{ type: "verify" }] },
        { identification: "username" },
      ],
    },
    { type: "fill_form" },
  ],
};

describe("layoutGraph", () => {
  it("reserves no extra room around an optional step", () => {
    const { lanes } = layout(optionalInBranch);
    expect([...lanes.keys()].filter((k) => k.startsWith("skip:"))).toEqual([]);
  });

  it("keeps every lane and row off the cards it passes", () => {
    for (const flow of [
      branchy,
      optionalInBranch,
      only(generatedEmailOAuthPasskey2FA.login_flows),
      only(generatedEmailOAuthPasskey2FA.signup_login_flows),
      only(generatedUsernameOnly.account_recovery_flows),
    ]) {
      expectEdgesClearOfCards(layout(flow));
    }
  });

  it("stacks the flows a step hands off to in the last column", () => {
    const { positions } = layout(
      only(generatedEmailOAuthPasskey2FA.signup_login_flows)
    );
    // Two terminals: 100 + 40 + 100 tall, centred on the row.
    expect(at(positions, flowRefNodeID("login", "default"))).toEqual({
      x: 2 * COLUMN,
      y: -120,
    });
    expect(at(positions, flowRefNodeID("signup", "default"))).toEqual({
      x: 2 * COLUMN,
      y: 20,
    });
  });

  it("does not overlap nodes in the generated login flow", () => {
    const { graph, positions } = layout(
      only(generatedEmailOAuthPasskey2FA.login_flows)
    );
    expectNoOverlap(graph, positions);
  });

  it("places steps after an empty choice", () => {
    const { graph, positions } = layout(
      only(generatedUsernameOnly.account_recovery_flows)
    );
    expectNoOverlap(graph, positions);
    expect(at(positions, "/steps/1").x).toBeGreaterThan(
      at(positions, "/steps/0").x
    );
  });

  it("throws when a node has no size", () => {
    const tree = flowToTree(branchy);
    expect(() => layoutGraph(tree, new Map(), [END_NODE_ID])).toThrow(
      /no size/
    );
  });
});

describe("layoutGraph with boxed choices", () => {
  function boxed(flow: AuthenticationFlowConfig) {
    const tree = flowToTree(flow);
    const graph = treeToGraph(tree);
    return {
      graph,
      ...layoutGraph(tree, uniformSizes(graph), graph.terminalIDs),
    };
  }

  it("boxes the choice and its options, one row per option", () => {
    const { boxes, positions } = boxed(branchy);
    const cardX = COLUMN + BOX_PAD;
    const branchX = cardX + COLUMN;
    // Email's two steps are the widest option.
    const right = branchX + COLUMN + W + BOX_EXIT;
    const height = 3 * GAP_Y + 2 * H + 32;
    expect(boxes).toEqual([
      {
        id: "/steps/0",
        x: COLUMN,
        y: -height / 2,
        width: right - COLUMN,
        height,
        rowsX: cardX + W + GAP_X / 2,
        rows: [
          { optionIndex: 0, top: -height / 2, bottom: -36 },
          { optionIndex: 1, top: -36, bottom: 104 },
          { optionIndex: 2, top: 104, bottom: height / 2 },
        ],
      },
    ]);
    expect(at(positions, "/steps/0")).toEqual({ x: cardX, y: -50 });
    expect(at(positions, "/steps/0/one_of/0/steps/0")).toEqual({
      x: branchX,
      y: -156,
    });
    expect(at(positions, joinNodeID("/steps/0"))).toEqual({
      x: right - JOIN / 2,
      y: -JOIN / 2,
    });
    expect(at(positions, "/steps/1")).toEqual({ x: right + GAP_X, y: -50 });
  });

  it("reserves a lane through the middle of every row", () => {
    const { lanes } = boxed(branchy);
    expect(
      [0, 1, 2].map((j) => lanes.get(optionLaneKey("/steps/0", j)))
    ).toEqual([-106, 34, 140]);
  });

  it("nests a box inside its option's row", () => {
    const { boxes } = boxed({
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [
                {
                  type: "create_authenticator",
                  one_of: [
                    {
                      authentication: "primary_password",
                      steps: [{ type: "verify" }],
                    },
                    { authentication: "primary_passkey" },
                  ],
                },
              ],
            },
            { identification: "username" },
          ],
        },
      ],
    });
    const outer = boxes.find((b) => b.id === "/steps/0");
    const inner = boxes.find((b) => b.id === "/steps/0/one_of/0/steps/0");
    if (outer == null || inner == null) {
      throw new Error("expected two boxes");
    }
    const row = outer.rows[0];
    expect(inner.x).toBeGreaterThanOrEqual(outer.rowsX);
    expect(inner.x + inner.width).toBeLessThanOrEqual(outer.x + outer.width);
    expect(inner.y).toBeGreaterThanOrEqual(row.top);
    expect(inner.y + inner.height).toBeLessThanOrEqual(row.bottom);
  });
});
