import { describe, it, expect } from "@jest/globals";
import { flowToTree } from "./tree";
import {
  END_NODE_ID,
  FlowGraph,
  NEXT_HANDLE_ID,
  START_NODE_ID,
  flowRefNodeID,
  handoffGroups,
  optionAnswerHandleID,
  optionHandleID,
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
      "/steps/1",
      "/steps/2",
      "/steps/3",
      END_NODE_ID,
    ]);
    expect(edgeTuples(graph)).toEqual(
      expect.arrayContaining([
        [START_NODE_ID, NEXT_HANDLE_ID, "next", "/steps/0"],
        ["/steps/0", optionHandleID(0), "option", "/steps/1"],
        ["/steps/0", optionHandleID(1), "option", P],
        ["/steps/0", optionHandleID(2), "option", "/steps/1"],
        ["/steps/0", optionHandleID(3), "option", "/steps/1"],
        [P, optionHandleID(0), "option", S],
        [S, optionHandleID(0), "option", C],
        [S, optionHandleID(1), "option", C],
        [S, optionHandleID(2), "option", C],
        [C, NEXT_HANDLE_ID, "next", "/steps/1"],
        [P, optionHandleID(1), "option", "/steps/1"],
        ["/steps/1", NEXT_HANDLE_ID, "next", "/steps/2"],
        ["/steps/2", NEXT_HANDLE_ID, "next", "/steps/3"],
        ["/steps/3", NEXT_HANDLE_ID, "next", END_NODE_ID],
      ])
    );
    expect(graph.edges).toHaveLength(14);
    expect(new Set(graph.edges.map((e) => e.id)).size).toEqual(14);
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
