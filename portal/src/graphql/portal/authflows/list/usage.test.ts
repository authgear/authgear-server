import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import { flowUses, removeFlowUses, renameFlowUses } from "./usage";

const config = {
  ui: {
    authentication_flow: {
      groups: [
        {
          name: "staff",
          flows: [
            { type: "signup", name: "shared_2fa" },
            { type: "login", name: "default" },
          ],
        },
      ],
    },
  },
  oauth: {
    clients: [
      {
        client_id: "c1",
        name: "Admin app",
        x_authentication_flow_allowlist: {
          flows: [{ type: "signup", name: "shared_2fa" }],
        },
      },
      { client_id: "c2", name: "Other" },
    ],
  },
} as unknown as PortalAPIAppConfig;

describe("flowUses", () => {
  it("lists the groups and applications that name the flow", () => {
    expect(flowUses(config, "signup", "shared_2fa")).toEqual([
      { kind: "group", name: "staff" },
      { kind: "application", name: "Admin app" },
    ]);
  });

  it("matches the type as well as the name", () => {
    expect(flowUses(config, "login", "shared_2fa")).toEqual([]);
  });

  it("never reports default: the generated flow always answers to it", () => {
    expect(flowUses(config, "login", "default")).toEqual([]);
  });
});

describe("renameFlowUses", () => {
  it("points groups and applications at the new name", () => {
    const next = renameFlowUses(config, "signup", "shared_2fa", "two_factor");
    expect(flowUses(next, "signup", "two_factor")).toEqual([
      { kind: "group", name: "staff" },
      { kind: "application", name: "Admin app" },
    ]);
    expect(flowUses(next, "signup", "shared_2fa")).toEqual([]);
  });
});

describe("removeFlowUses", () => {
  it("falls groups back to default and drops the flow from applications", () => {
    const next = removeFlowUses(config, "signup", "shared_2fa") as unknown as {
      ui: { authentication_flow: { groups: { flows: unknown[] }[] } };
      oauth: {
        clients: { x_authentication_flow_allowlist?: { flows: unknown[] } }[];
      };
    };
    expect(next.ui.authentication_flow.groups[0].flows).toEqual([
      { type: "signup", name: "default" },
      { type: "login", name: "default" },
    ]);
    expect(
      next.oauth.clients[0].x_authentication_flow_allowlist?.flows
    ).toEqual([]);
    expect(
      flowUses(next as unknown as PortalAPIAppConfig, "signup", "shared_2fa")
    ).toEqual([]);
  });
});
