import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import {
  applyGroupAction,
  flowGroups,
  flowNamesOfType,
  groupDeleteBlockers,
  groupNameError,
  groupSelection,
  groupUsers,
  standardGroupUsers,
} from "./groups";

const config: PortalAPIAppConfig = {
  id: "app",
  authentication_flow: {
    login_flows: [{ name: "default" }, { name: "staff" }],
    signup_flows: [{ name: "staff" }],
  },
  ui: {
    authentication_flow: {
      groups: [
        {
          name: "staff",
          flows: [
            { type: "login", name: "staff" },
            { type: "signup", name: "staff" },
            { type: "signup", name: "extra" },
          ],
        },
        { name: "public", flows: [{ type: "login", name: "default" }] },
      ],
    },
  },
  oauth: {
    clients: [
      {
        client_id: "c1",
        name: "Admin app",
        x_authentication_flow_allowlist: { groups: [{ name: "staff" }] },
      },
      {
        client_id: "c2",
        x_authentication_flow_allowlist: {
          groups: [{ name: "staff" }, { name: "public" }],
        },
      },
      {
        client_id: "c3",
        x_authentication_flow_allowlist: {
          groups: [{ name: "staff" }],
          flows: [{ type: "login", name: "staff" }],
        },
      },
      { client_id: "c4" },
    ],
  },
};

describe("flowNamesOfType", () => {
  it("lists default first, then the flows of the type", () => {
    expect(flowNamesOfType(config.authentication_flow, "login")).toEqual([
      "default",
      "staff",
    ]);
    expect(flowNamesOfType(config.authentication_flow, "reauth")).toEqual([
      "default",
    ]);
  });
});

describe("groupSelection", () => {
  it("takes the first flow of each type, and default for a missing one", () => {
    const [staff] = flowGroups(config);
    expect(groupSelection(staff)).toEqual({
      signup_login: "default",
      signup: "staff",
      login: "staff",
      promote: "default",
      reauth: "default",
      account_recovery: "default",
    });
  });
});

describe("groupUsers", () => {
  it("names the applications that allow the group", () => {
    expect(groupUsers(config, "staff")).toEqual(["Admin app", "c2", "c3"]);
    expect(groupUsers(config, "public")).toEqual(["c2"]);
  });
});

describe("groupDeleteBlockers", () => {
  it("names the applications that allow nothing but the group", () => {
    expect(groupDeleteBlockers(config, "staff")).toEqual(["Admin app"]);
    expect(groupDeleteBlockers(config, "public")).toEqual([]);
  });
});

describe("groupNameError", () => {
  it("checks the pattern, the reserved name, and other groups", () => {
    expect(groupNameError(config, "new_group", null)).toBeNull();
    expect(groupNameError(config, "1bad", null)).toBe("invalid");
    expect(groupNameError(config, "default", null)).toBe("reserved");
    expect(groupNameError(config, "public", null)).toBe("taken");
    expect(groupNameError(config, "public", "public")).toBeNull();
  });
});

describe("applyGroupAction", () => {
  it("adds a group with a flow of every type", () => {
    const next = applyGroupAction(config, {
      kind: "add",
      name: "kiosk",
      selection: {
        signup_login: "default",
        signup: "staff",
        login: "staff",
        promote: "default",
        reauth: "default",
        account_recovery: "default",
      },
    });
    expect(next.ui?.authentication_flow?.groups?.[2]).toEqual({
      name: "kiosk",
      flows: [
        { type: "signup_login", name: "default" },
        { type: "signup", name: "staff" },
        { type: "login", name: "staff" },
        { type: "promote", name: "default" },
        { type: "reauth", name: "default" },
        { type: "account_recovery", name: "default" },
      ],
    });
  });

  it("adds the first group to a config without any", () => {
    const next = applyGroupAction(
      { id: "app" },
      { kind: "add", name: "g", selection: groupSelection(null) }
    );
    expect(next.ui?.authentication_flow?.groups?.[0].name).toBe("g");
  });

  it("sets a group's flows, keeping extra flows of a type", () => {
    const next = applyGroupAction(config, {
      kind: "set-flows",
      name: "staff",
      selection: { ...groupSelection(null), signup: "default" },
    });
    expect(next.ui?.authentication_flow?.groups?.[0].flows).toEqual([
      { type: "login", name: "default" },
      { type: "signup", name: "default" },
      { type: "signup", name: "extra" },
      { type: "signup_login", name: "default" },
      { type: "promote", name: "default" },
      { type: "reauth", name: "default" },
      { type: "account_recovery", name: "default" },
    ]);
  });

  it("renames a group and the applications that allow it", () => {
    const next = applyGroupAction(config, {
      kind: "rename",
      name: "staff",
      newName: "employees",
    });
    expect(next.ui?.authentication_flow?.groups?.[0].name).toBe("employees");
    expect(
      next.oauth?.clients?.[1].x_authentication_flow_allowlist?.groups
    ).toEqual([{ name: "employees" }, { name: "public" }]);
  });

  it("deletes a group and lets the applications go of it", () => {
    const next = applyGroupAction(config, { kind: "delete", name: "public" });
    expect(next.ui?.authentication_flow?.groups?.map((g) => g.name)).toEqual([
      "staff",
    ]);
    expect(
      next.oauth?.clients?.[1].x_authentication_flow_allowlist?.groups
    ).toEqual([{ name: "staff" }]);
  });
});

describe("standardGroupUsers", () => {
  it("names the applications with no groups, or that list the default group", () => {
    const withDefault = {
      ...config,
      oauth: {
        clients: [
          ...(config.oauth?.clients ?? []),
          {
            client_id: "c5",
            x_authentication_flow_allowlist: {
              groups: [{ name: "default" }],
            },
          },
          { client_id: "m2m", x_application_type: "m2m" as const },
        ],
      },
    };
    expect(standardGroupUsers(withDefault)).toEqual(["c4", "c5"]);
  });
});
