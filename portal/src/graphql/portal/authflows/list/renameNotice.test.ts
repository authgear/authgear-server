import { describe, expect, it } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import { renameNotice } from "./usage";

const render = (id: string, values?: Record<string, string>) =>
  `${id}(${values?.names ?? ""})`;

const config = {
  id: "app",
  authentication_flow: {
    login_flows: [{ name: "staff" }],
    signup_login_flows: [
      {
        name: "employee",
        steps: [
          {
            type: "identify",
            one_of: [{ identification: "email", login_flow: "staff" }],
          },
        ],
      },
    ],
  },
  ui: {
    authentication_flow: {
      groups: [{ name: "employee", flows: [{ type: "login", name: "staff" }] }],
    },
  },
  oauth: {
    clients: [
      {
        client_id: "c1",
        name: "Staff Portal",
        x_authentication_flow_allowlist: {
          flows: [{ type: "login", name: "staff" }],
        },
      },
    ],
  },
} as unknown as PortalAPIAppConfig;

describe("renameNotice", () => {
  it("names the groups, applications and hand-offs that follow the rename", () => {
    expect(renameNotice(render, config, "login", "staff")).toEqual(
      [
        "AuthFlowsScreen.list.rename.effect.groups(employee)",
        "AuthFlowsScreen.list.rename.effect.applications(Staff Portal)",
        "AuthFlowsScreen.list.rename.effect.handoffs(employee)",
      ].join(" ")
    );
  });

  it("says nothing for a flow nothing else names", () => {
    expect(renameNotice(render, config, "signup", "staff")).toBeNull();
    expect(renameNotice(render, null, "login", "staff")).toBeNull();
  });
});
