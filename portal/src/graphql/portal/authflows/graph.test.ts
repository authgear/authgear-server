import { describe, it, expect } from "@jest/globals";
import { flowToTree } from "./tree";
import {
  END_NODE_ID,
  FlowGraph,
  NEXT_HANDLE_ID,
  START_NODE_ID,
  flowRefNodeID,
  handoffGroups,
  joinNodeID,
  optionAnswerHandleID,
  optionHandleID,
  optionLaneKey,
  recoveryDestinationHandleID,
  treeToGraph,
} from "./graph";
import {
  customSignupWithForm,
  generatedEmailOAuthPasskey2FA,
  generatedUsernameOnly,
} from "./__fixtures__/flows";
import { AuthenticationFlowConfig } from "../../../types";

function only<T>(xs: T[] | undefined): T {
  if (xs?.length !== 1) {
    throw new Error("expected exactly one element");
  }
  return xs[0];
}

function graphOf(flow: AuthenticationFlowConfig): FlowGraph {
  return treeToGraph(flowToTree(flow));
}

function edgeTuples(graph: FlowGraph) {
  return graph.edges.map((e) => [e.source, e.sourceHandle, e.kind, e.target]);
}

const P = "/steps/0/one_of/1/steps/0"; // authenticate_primary_email
const S = `${P}/one_of/0/steps/0`; // optional 2FA
const C = `${P}/one_of/0/steps/1`; // change_password

describe("treeToGraph", () => {
  it("connects the generated login flow", () => {
    const graph = graphOf(only(generatedEmailOAuthPasskey2FA.login_flows));
    expect(graph.nodes.map((n) => n.id)).toEqual([
      START_NODE_ID,
      "/steps/0",
      P,
      S,
      C,
      joinNodeID(P),
      joinNodeID("/steps/0"),
      "/steps/1",
      "/steps/2",
      "/steps/3",
      END_NODE_ID,
    ]);
    // Both choices with steps under an option are boxed: their options meet
    // at the box's join before what follows.
    expect(edgeTuples(graph)).toEqual(
      expect.arrayContaining([
        [START_NODE_ID, NEXT_HANDLE_ID, "next", "/steps/0"],
        ["/steps/0", optionHandleID(0), "option", joinNodeID("/steps/0")],
        ["/steps/0", optionHandleID(1), "option", P],
        ["/steps/0", optionHandleID(2), "option", joinNodeID("/steps/0")],
        ["/steps/0", optionHandleID(3), "option", joinNodeID("/steps/0")],
        [P, optionHandleID(0), "option", S],
        [S, optionHandleID(0), "option", C],
        [S, optionHandleID(1), "option", C],
        [S, optionHandleID(2), "option", C],
        [C, NEXT_HANDLE_ID, "next", joinNodeID(P)],
        [P, optionHandleID(1), "option", joinNodeID(P)],
        [joinNodeID(P), NEXT_HANDLE_ID, "next", joinNodeID("/steps/0")],
        [joinNodeID("/steps/0"), NEXT_HANDLE_ID, "next", "/steps/1"],
        ["/steps/1", NEXT_HANDLE_ID, "next", "/steps/2"],
        ["/steps/2", NEXT_HANDLE_ID, "next", "/steps/3"],
        ["/steps/3", NEXT_HANDLE_ID, "next", END_NODE_ID],
      ])
    );
    expect(graph.edges).toHaveLength(16);
    expect(new Set(graph.edges.map((e) => e.id)).size).toEqual(16);
  });

  it("names the lane each option edge travels in", () => {
    const graph = graphOf(only(generatedEmailOAuthPasskey2FA.login_flows));
    const lane = (source: string, handle: string, kind: string) =>
      graph.edges.find(
        (e) =>
          e.source === source && e.sourceHandle === handle && e.kind === kind
      )?.lane;
    expect(lane("/steps/0", optionHandleID(2), "option")).toEqual(
      `/steps/0/${optionHandleID(2)}`
    );
    expect(lane("/steps/1", NEXT_HANDLE_ID, "next")).toBeNull();
  });

  it("draws no bypass around an optional step: its condition line says so", () => {
    const graph = graphOf(only(generatedEmailOAuthPasskey2FA.login_flows));
    expect(
      graph.edges.filter((e) => e.target === C).map((e) => e.source)
    ).toEqual([S, S, S]);
  });

  it("chains inlined single-option steps with next edges", () => {
    const graph = graphOf(only(customSignupWithForm.signup_flows));
    expect(edgeTuples(graph)).toEqual([
      [START_NODE_ID, NEXT_HANDLE_ID, "next", "/steps/0"],
      ["/steps/0", NEXT_HANDLE_ID, "next", "/steps/0/one_of/0/steps/0"],
      [
        "/steps/0/one_of/0/steps/0",
        NEXT_HANDLE_ID,
        "next",
        "/steps/0/one_of/0/steps/1",
      ],
      ["/steps/0/one_of/0/steps/1", NEXT_HANDLE_ID, "next", "/steps/1"],
      ["/steps/1", NEXT_HANDLE_ID, "next", END_NODE_ID],
    ]);
  });

  it("leaves an empty choice without outgoing edges", () => {
    const graph = graphOf(only(generatedUsernameOnly.account_recovery_flows));
    expect(graph.edges.filter((e) => e.source === "/steps/0")).toEqual([]);
    expect(graph.edges.filter((e) => e.target === "/steps/1")).toEqual([]);
    expect(graph.nodes.map((n) => n.id)).toContain("/steps/1");
  });

  it("hands signup_login options off to the flows they name, without an End", () => {
    const graph = graphOf(
      only(generatedEmailOAuthPasskey2FA.signup_login_flows)
    );
    const login = flowRefNodeID("login", "default");
    const signup = flowRefNodeID("signup", "default");
    expect(graph.nodes.map((n) => n.id)).toEqual([
      START_NODE_ID,
      "/steps/0",
      login,
      signup,
    ]);
    expect(edgeTuples(graph)).toEqual([
      [START_NODE_ID, NEXT_HANDLE_ID, "next", "/steps/0"],
      // Continue as and passkey only log in, from the option's own handle.
      ["/steps/0", optionHandleID(0), "option", login],
      // Email and Google hand off the same way, so they share one pair,
      // from Email, the first of the group.
      ["/steps/0", optionAnswerHandleID(1, "has_account"), "option", login],
      ["/steps/0", optionAnswerHandleID(1, "no_account"), "option", signup],
      ["/steps/0", optionHandleID(3), "option", login],
    ]);
    expect(graph.terminalIDs).toEqual([login, signup]);
  });

  const emailPhoneUsername: AuthenticationFlowConfig = {
    name: "default",
    steps: [
      {
        type: "identify",
        one_of: [
          { identification: "select_account", login_flow: "default" },
          {
            identification: "email",
            login_flow: "default",
            signup_flow: "default",
          },
          { identification: "passkey", login_flow: "default" },
          {
            identification: "phone",
            login_flow: "default",
            signup_flow: "default",
          },
          {
            identification: "username",
            login_flow: "default",
            signup_flow: "vip",
          },
          {
            identification: "oauth",
            login_flow: "default",
            signup_flow: "default",
          },
        ],
      },
    ],
  };

  it("groups options that hand off to the same two flows", () => {
    const options = emailPhoneUsername.steps![0].one_of!;
    expect(handoffGroups(options)).toEqual([[1, 3, 5], [4]]);
  });

  it("draws one has-account / no-account pair per hand-off group", () => {
    const graph = graphOf(emailPhoneUsername);
    const login = flowRefNodeID("login", "default");
    expect(edgeTuples(graph)).toEqual([
      [START_NODE_ID, NEXT_HANDLE_ID, "next", "/steps/0"],
      ["/steps/0", optionHandleID(0), "option", login],
      ["/steps/0", optionAnswerHandleID(1, "has_account"), "option", login],
      [
        "/steps/0",
        optionAnswerHandleID(1, "no_account"),
        "option",
        flowRefNodeID("signup", "default"),
      ],
      ["/steps/0", optionHandleID(2), "option", login],
      ["/steps/0", optionAnswerHandleID(4, "has_account"), "option", login],
      [
        "/steps/0",
        optionAnswerHandleID(4, "no_account"),
        "option",
        flowRefNodeID("signup", "vip"),
      ],
    ]);
  });

  it("lists End as the terminal of an ordinary flow", () => {
    const graph = graphOf(only(customSignupWithForm.signup_flows));
    expect(graph.terminalIDs).toEqual([END_NODE_ID]);
  });
});

describe("account recovery", () => {
  const recovery: AuthenticationFlowConfig = {
    name: "default",
    steps: [
      {
        type: "identify",
        one_of: [
          {
            identification: "email",
            steps: [
              {
                type: "select_destination",
                allowed_channels: [{ channel: "email", otp_form: "link" }],
              },
            ],
          },
          {
            identification: "phone",
            steps: [
              {
                type: "select_destination",
                allowed_channels: [{ channel: "sms", otp_form: "code" }],
              },
            ],
          },
        ],
      },
      { type: "verify_account_recovery_code" },
    ],
  };

  it("lands each destination on its own section of the verify step", () => {
    const graph = graphOf(recovery);
    const into = graph.edges
      .filter((e) => e.target === "/steps/1")
      .map((e) => [e.source, e.targetHandle]);
    expect(into).toEqual([
      [
        "/steps/0/one_of/0/steps/0",
        recoveryDestinationHandleID({ channel: "email", otp_form: "link" }),
      ],
      [
        "/steps/0/one_of/1/steps/0",
        recoveryDestinationHandleID({ channel: "sms", otp_form: "code" }),
      ],
    ]);
  });

  it("leaves other edges on the default target handle", () => {
    const graph = graphOf(recovery);
    expect(
      graph.edges
        .filter((e) => e.target !== "/steps/1")
        .map((e) => e.targetHandle)
    ).toEqual(
      graph.edges.filter((e) => e.target !== "/steps/1").map(() => null)
    );
  });
});

describe("insertion points", () => {
  const insertAtOf = (flow: AuthenticationFlowConfig) =>
    new Map(
      graphOf(flow).edges.map((e) => [`${e.source}->${e.target}`, e.insertAt])
    );

  it("puts each line kind at its config location", () => {
    const at = insertAtOf({
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [
                {
                  type: "authenticate",
                  one_of: [{ authentication: "primary_password" }],
                },
              ],
            },
            { identification: "phone" },
          ],
        },
        { type: "check_account_status" },
      ],
    });
    expect(at.get(`${START_NODE_ID}->/steps/0`)).toEqual({
      steps: "/steps",
      index: 0,
    });
    expect(at.get("/steps/0->/steps/0/one_of/0/steps/0")).toEqual({
      steps: "/steps/0/one_of/0/steps",
      index: 0,
    });
    expect(at.get("/steps/0/one_of/0/steps/0->/steps/0:join")).toEqual({
      steps: "/steps/0/one_of/0/steps",
      index: 1,
    });
    expect(at.get("/steps/0->/steps/0:join")).toEqual({
      steps: "/steps/0/one_of/1/steps",
      index: 0,
    });
    expect(at.get("/steps/0:join->/steps/1")).toEqual({
      steps: "/steps",
      index: 1,
    });
    expect(at.get(`/steps/1->${END_NODE_ID}`)).toEqual({
      steps: "/steps",
      index: 2,
    });
  });

  it("inserts after an inline single-option step into its nested steps", () => {
    const at = insertAtOf({
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [{ type: "check_account_status" }],
            },
          ],
        },
      ],
    });
    expect(at.get("/steps/0->/steps/0/one_of/0/steps/0")).toEqual({
      steps: "/steps/0/one_of/0/steps",
      index: 0,
    });
  });

  it("inserts after an inline step with no nested steps in its own list", () => {
    const at = insertAtOf({
      name: "default",
      steps: [
        { type: "identify", one_of: [{ identification: "email" }] },
        { type: "check_account_status" },
      ],
    });
    expect(at.get("/steps/0->/steps/1")).toEqual({ steps: "/steps", index: 1 });
  });

  it("has no insertion point on a hand-off line", () => {
    const graph = graphOf({
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              login_flow: "default",
              signup_flow: "default",
            },
            {
              identification: "oauth",
              login_flow: "default",
              signup_flow: "default",
            },
          ],
        },
      ],
    });
    const handoffs = graph.edges.filter((e) => e.target.startsWith("flow:"));
    expect(handoffs.length).toBeGreaterThan(0);
    expect(handoffs.every((e) => e.insertAt === null)).toBe(true);
  });
});

describe("insertion points after a hand-off", () => {
  it("has none after a single inline hand-off option", () => {
    const graph = graphOf({
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              login_flow: "default",
              signup_flow: "default",
            },
          ],
        },
      ],
    });
    expect(
      graph.edges.filter((e) => e.source === "/steps/0").map((e) => e.insertAt)
    ).toEqual(
      graph.edges.filter((e) => e.source === "/steps/0").map(() => null)
    );
  });
});

// A choice whose options have no steps is not boxed; its lines meet at what
// follows, and the first of them carries the point after the choice.
describe("insertion after a choice", () => {
  const flow: AuthenticationFlowConfig = {
    name: "default",
    steps: [
      {
        type: "identify",
        one_of: [{ identification: "email" }, { identification: "phone" }],
      },
    ],
  };

  it("marks one line where the options meet with the point after the choice", () => {
    const merges = graphOf(flow)
      .edges.filter((e) => e.mergeInsertAt != null)
      .map((e) => [e.target, e.mergeInsertAt]);
    expect(merges).toEqual([[END_NODE_ID, { steps: "/steps", index: 1 }]]);
  });

  it("does so before the next step, not only before End", () => {
    const merges = graphOf({
      ...flow,
      steps: [...(flow.steps ?? []), { type: "view_recovery_code" }],
    })
      .edges.filter((e) => e.mergeInsertAt != null)
      .map((e) => [e.target, e.mergeInsertAt]);
    expect(merges).toEqual([["/steps/1", { steps: "/steps", index: 1 }]]);
  });

  it("has none after a step that is not a choice", () => {
    expect(
      graphOf({
        name: "default",
        steps: [{ type: "check_account_status" }],
      }).edges.filter((e) => e.mergeInsertAt != null)
    ).toEqual([]);
  });
});

describe("boxed choices", () => {
  const flow: AuthenticationFlowConfig = {
    name: "default",
    steps: [
      {
        type: "identify",
        one_of: [
          { identification: "email", steps: [{ type: "verify" }] },
          { identification: "username" },
        ],
      },
      { type: "fill_form" },
    ],
  };
  const join = joinNodeID("/steps/0");

  it("joins every option at the box's exit, then leads on once", () => {
    const graph = treeToGraph(flowToTree(flow));
    expect(graph.nodes.map((n) => [n.id, n.kind])).toEqual([
      [START_NODE_ID, "start"],
      ["/steps/0", "page"],
      ["/steps/0/one_of/0/steps/0", "page"],
      [join, "join"],
      ["/steps/1", "page"],
      [END_NODE_ID, "end"],
    ]);
    expect(edgeTuples(graph)).toEqual([
      [START_NODE_ID, NEXT_HANDLE_ID, "next", "/steps/0"],
      ["/steps/0", optionHandleID(0), "option", "/steps/0/one_of/0/steps/0"],
      ["/steps/0/one_of/0/steps/0", NEXT_HANDLE_ID, "next", join],
      ["/steps/0", optionHandleID(1), "option", join],
      [join, NEXT_HANDLE_ID, "next", "/steps/1"],
      ["/steps/1", NEXT_HANDLE_ID, "next", END_NODE_ID],
    ]);
  });

  it("adds after the choice from the exit line, and nowhere else", () => {
    const graph = treeToGraph(flowToTree(flow));
    const exit = graph.edges.find((e) => e.source === join);
    expect(exit?.insertAt).toEqual({ steps: "/steps", index: 1 });
    expect(exit?.afterChoice).toEqual(true);
    expect(graph.edges.filter((e) => e.mergeInsertAt != null)).toEqual([]);
  });

  it("runs each option's line to the exit along its row", () => {
    const graph = treeToGraph(flowToTree(flow));
    const into = graph.edges.filter((e) => e.target === join);
    expect(into.map((e) => [e.source, e.lane])).toEqual([
      ["/steps/0/one_of/0/steps/0", optionLaneKey("/steps/0", 0)],
      ["/steps/0", optionLaneKey("/steps/0", 1)],
    ]);
  });

  it("draws no box for a choice whose options have no steps", () => {
    const graph = treeToGraph(
      flowToTree({
        name: "default",
        steps: [
          {
            type: "identify",
            one_of: [
              { identification: "email" },
              { identification: "username" },
            ],
          },
        ],
      })
    );
    expect(graph.nodes.some((n) => n.kind === "join")).toEqual(false);
  });

  it("keeps account recovery's lines to each destination's input", () => {
    const recovery: AuthenticationFlowConfig = {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [
                {
                  type: "select_destination",
                  allowed_channels: [{ channel: "email", otp_form: "link" }],
                },
              ],
            },
            {
              identification: "phone",
              steps: [
                {
                  type: "select_destination",
                  allowed_channels: [{ channel: "sms", otp_form: "code" }],
                },
              ],
            },
          ],
        },
        { type: "verify_account_recovery_code" },
      ],
    };
    const graph = treeToGraph(flowToTree(recovery));
    expect(graph.nodes.some((n) => n.kind === "join")).toEqual(false);
    expect(
      graph.edges
        .filter((e) => e.target === "/steps/1")
        .map((e) => [e.source, e.targetHandle])
    ).toEqual([
      [
        "/steps/0/one_of/0/steps/0",
        recoveryDestinationHandleID({ channel: "email", otp_form: "link" }),
      ],
      [
        "/steps/0/one_of/1/steps/0",
        recoveryDestinationHandleID({ channel: "sms", otp_form: "code" }),
      ],
    ]);
  });
});
