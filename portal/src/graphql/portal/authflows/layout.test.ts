import { describe, it, expect } from "@jest/globals";
import { flowToTree } from "./tree";
import {
  END_NODE_ID,
  FlowGraph,
  START_NODE_ID,
  flowRefNodeID,
  treeToGraph,
} from "./graph";
import { GAP_X, Position, Size, layoutGraph } from "./layout";
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

function uniformSizes(graph: FlowGraph): Map<string, Size> {
  return new Map(graph.nodes.map((n) => [n.id, { width: W, height: H }]));
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
  const rects = graph.nodes.map((n) => ({ id: n.id, ...at(positions, n.id) }));
  for (let i = 0; i < rects.length; i++) {
    for (let j = i + 1; j < rects.length; j++) {
      const a = rects[i];
      const b = rects[j];
      const overlaps =
        a.x < b.x + W && b.x < a.x + W && a.y < b.y + H && b.y < a.y + H;
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
    const s = at(positions, e.source);
    const t = at(positions, e.target);
    const from = s.x + W + GAP_X / 2;
    const to = t.x - GAP_X / 2;
    if (to - from <= 0) {
      continue;
    }
    const laneY = e.lane != null ? lanes.get(e.lane) : undefined;
    if (laneY == null && e.kind !== "next") {
      throw new Error(`${e.id} skips columns without a lane`);
    }
    const y = laneY ?? s.y + H / 2;
    for (const n of graph.nodes) {
      if (n.id === e.source || n.id === e.target) {
        continue;
      }
      const r = at(positions, n.id);
      if (r.x < to && r.x + W > from && r.y < y && y < r.y + H) {
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
  it("places branches in config order, centred on the choice", () => {
    const { positions } = layout(branchy);
    expect(at(positions, START_NODE_ID)).toEqual({ x: 0, y: -50 });
    expect(at(positions, "/steps/0")).toEqual({ x: COLUMN, y: -50 });
    expect(at(positions, "/steps/0/one_of/0/steps/0")).toEqual({
      x: 2 * COLUMN,
      y: -156,
    });
    expect(at(positions, "/steps/0/one_of/0/steps/1")).toEqual({
      x: 3 * COLUMN,
      y: -156,
    });
    expect(at(positions, "/steps/0/one_of/1/steps/0")).toEqual({
      x: 2 * COLUMN,
      y: -16,
    });
  });

  it("reserves a lane below the branches for an empty branch", () => {
    const { lanes } = layout(branchy);
    expect(lanes.get("/steps/0/option-2")).toEqual(140);
  });

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

  it("places the join right of the widest branch", () => {
    const { positions } = layout(branchy);
    expect(at(positions, "/steps/1")).toEqual({ x: 4 * COLUMN, y: -50 });
    expect(at(positions, END_NODE_ID)).toEqual({ x: 5 * COLUMN, y: -50 });
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
