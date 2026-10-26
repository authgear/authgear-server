import { describe, expect, it } from "@jest/globals";
import {
  AuthenticationFlowConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import {
  authenticatedBefore,
  checkAccountStatusAt,
  fillablePointers,
  knownPointers,
  flowChecks,
  identifiedBefore,
  stepChecks,
} from "./checks";

const form = (...pointers: string[]) => ({
  fields: pointers.map((pointer) => ({ user_profile: { pointer } })),
});

describe("fillablePointers", () => {
  it("keeps attributes the end user can read and write", () => {
    const config = {
      id: "app",
      user_profile: {
        standard_attributes: {
          access_control: [
            {
              pointer: "/given_name",
              access_control: {
                end_user: "readwrite",
                bearer: "readonly",
                portal_ui: "readwrite",
              },
            },
            {
              pointer: "/gender",
              access_control: {
                end_user: "readonly",
                bearer: "readonly",
                portal_ui: "readwrite",
              },
            },
            {
              pointer: "/email",
              access_control: {
                end_user: "readwrite",
                bearer: "readonly",
                portal_ui: "readwrite",
              },
            },
            {
              pointer: "/address",
              access_control: {
                end_user: "readwrite",
                bearer: "readonly",
                portal_ui: "readwrite",
              },
            },
          ],
        },
        custom_attributes: {
          attributes: [
            {
              id: "1",
              pointer: "/x_team",
              type: "string",
              access_control: {
                end_user: "readwrite",
                bearer: "readonly",
                portal_ui: "readwrite",
              },
            },
            {
              id: "2",
              pointer: "/x_tier",
              type: "string",
              access_control: {
                end_user: "hidden",
                bearer: "hidden",
                portal_ui: "readwrite",
              },
            },
          ],
        },
      },
    } as PortalAPIAppConfig;
    expect(fillablePointers(config)).toEqual([
      "/given_name",
      "/address",
      "/x_team",
    ]);
  });
});

const loginFlow: AuthenticationFlowConfig = {
  name: "f",
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
            { type: "fill_form", form: form("/given_name") },
          ],
        },
        {
          identification: "oauth",
          steps: [{ type: "fill_form", form: form("/given_name") }],
        },
        {
          identification: "username",
          steps: [{ type: "fill_form", form: form("/given_name") }],
        },
      ],
    },
    { type: "fill_form", form: form("/given_name") },
    { type: "check_account_status" },
  ],
};

describe("identifiedBefore", () => {
  it("follows the server: an earlier identify step, or an identify option", () => {
    const flow: AuthenticationFlowConfig = {
      name: "f",
      steps: [
        { type: "fill_form" },
        {
          type: "identify",
          one_of: [{ identification: "email", steps: [{ type: "fill_form" }] }],
        },
        { type: "fill_form" },
      ],
    };
    expect(identifiedBefore(flow, "/steps/0")).toBe(false);
    expect(identifiedBefore(flow, "/steps/1/one_of/0/steps/0")).toBe(true);
    expect(identifiedBefore(flow, "/steps/2")).toBe(true);
    // Where a step would be inserted counts the same way.
    expect(identifiedBefore(flow, "/steps/3")).toBe(true);
  });
});

describe("authenticatedBefore", () => {
  it("is true after an authenticate step on the path", () => {
    expect(authenticatedBefore(loginFlow, "/steps/0/one_of/0/steps/1")).toBe(
      true
    );
  });

  it("is true under an option that authenticates by itself", () => {
    expect(authenticatedBefore(loginFlow, "/steps/0/one_of/1/steps/0")).toBe(
      true
    );
  });

  it("is false under an option that has not authenticated yet", () => {
    expect(authenticatedBefore(loginFlow, "/steps/0/one_of/2/steps/0")).toBe(
      false
    );
  });

  it("needs every option of an earlier identify step to authenticate", () => {
    expect(authenticatedBefore(loginFlow, "/steps/1")).toBe(false);
    const allAuthenticate: AuthenticationFlowConfig = {
      name: "f",
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
            { identification: "passkey" },
          ],
        },
        { type: "fill_form" },
      ],
    };
    expect(authenticatedBefore(allAuthenticate, "/steps/1")).toBe(true);
  });
});

describe("stepChecks", () => {
  const fillable = new Set(["/given_name", "/family_name"]);

  it("finds a form before any identify step", () => {
    const flow: AuthenticationFlowConfig = {
      name: "f",
      steps: [{ type: "fill_form", form: form("/given_name") }],
    };
    expect(stepChecks("signup", flow, fillable).get("/steps/0")).toEqual([
      { kind: "form_before_identify" },
    ]);
  });

  it("finds a login form before the user is authenticated", () => {
    const checks = stepChecks("login", loginFlow, fillable);
    expect(checks.get("/steps/0/one_of/2/steps/0")).toEqual([
      { kind: "form_before_authentication" },
    ]);
    expect(checks.get("/steps/1")).toEqual([
      { kind: "form_before_authentication" },
    ]);
    expect(checks.has("/steps/0/one_of/0/steps/1")).toBe(false);
    expect(checks.has("/steps/0/one_of/1/steps/0")).toBe(false);
  });

  it("does not ask signup forms to come after authentication", () => {
    expect(stepChecks("signup", loginFlow, fillable).has("/steps/1")).toBe(
      false
    );
  });

  it("finds fields the end user cannot fill, and repeated fields", () => {
    const flow: AuthenticationFlowConfig = {
      name: "f",
      steps: [
        {
          type: "identify",
          one_of: [{ identification: "email", form: form("/gender") }],
        },
        {
          type: "fill_form",
          form: form("/given_name", "/email", "/given_name"),
        },
      ],
    };
    const checks = stepChecks("signup", flow, fillable);
    expect(checks.get("/steps/0")).toEqual([
      { kind: "field_not_fillable", pointer: "/gender", option: 0 },
    ]);
    expect(checks.get("/steps/1")).toEqual([
      { kind: "field_not_fillable", pointer: "/email" },
      { kind: "field_duplicate", pointer: "/given_name" },
    ]);
  });

  it("checks nothing in flows without forms", () => {
    expect(stepChecks("signup_login", loginFlow, fillable).size).toBe(0);
  });
});

describe("flowChecks", () => {
  it("warns when a login flow never checks the account status", () => {
    expect(flowChecks("login", loginFlow)).toEqual([]);
    const without: AuthenticationFlowConfig = {
      name: "f",
      steps: [{ type: "identify", one_of: [{ identification: "email" }] }],
    };
    expect(flowChecks("login", without)).toEqual(["no_check_account_status"]);
    expect(flowChecks("signup", without)).toEqual([]);
  });

  it("finds the step nested in an option", () => {
    const nested: AuthenticationFlowConfig = {
      name: "f",
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
    };
    expect(flowChecks("login", nested)).toEqual([]);
  });
});

describe("checkAccountStatusAt", () => {
  it("goes right after the last top-level identify or authenticate step", () => {
    const flow: AuthenticationFlowConfig = {
      name: "f",
      steps: [
        { type: "identify", one_of: [{ identification: "email" }] },
        {
          type: "authenticate",
          one_of: [{ authentication: "primary_password" }],
        },
        { type: "terminate_other_sessions" },
      ],
    };
    expect(checkAccountStatusAt(flow)).toEqual({ steps: "/steps", index: 2 });
  });

  it("goes at the start when there is neither", () => {
    expect(
      checkAccountStatusAt({ name: "f", steps: [{ type: "change_password" }] })
    ).toEqual({ steps: "/steps", index: 0 });
  });
});

describe("review fixes", () => {
  it("counts select_account as signed in, like generated login flows", () => {
    const flow: AuthenticationFlowConfig = {
      name: "f",
      steps: [
        {
          type: "identify",
          one_of: [
            { identification: "select_account" },
            {
              identification: "email",
              steps: [
                {
                  type: "authenticate",
                  one_of: [{ authentication: "primary_password" }],
                },
              ],
            },
            { identification: "oauth" },
          ],
        },
        { type: "check_account_status" },
        { type: "fill_form" },
      ],
    };
    expect(authenticatedBefore(flow, "/steps/2")).toBe(true);
  });

  it("lets a form ask for the whole address, as the server does", () => {
    const config = {
      id: "app",
      user_profile: {
        standard_attributes: {
          access_control: [
            {
              pointer: "/address",
              access_control: {
                end_user: "readwrite",
                bearer: "readonly",
                portal_ui: "readwrite",
              },
            },
          ],
        },
      },
    } as PortalAPIAppConfig;
    expect(fillablePointers(config)).toEqual(["/address"]);
  });

  it("lists checks the server rejects before warnings", () => {
    const flow: AuthenticationFlowConfig = {
      name: "f",
      steps: [
        { type: "identify", one_of: [{ identification: "email" }] },
        { type: "fill_form", form: form("/name") },
      ],
    };
    expect(stepChecks("login", flow, new Set()).get("/steps/1")).toEqual([
      { kind: "field_not_fillable", pointer: "/name" },
      { kind: "form_before_authentication" },
    ]);
  });
});

describe("knownPointers", () => {
  it("lists every attribute of the project, fillable or not", () => {
    const config = {
      id: "app",
      user_profile: {
        standard_attributes: {
          access_control: [
            {
              pointer: "/gender",
              access_control: {
                end_user: "hidden",
                bearer: "hidden",
                portal_ui: "readwrite",
              },
            },
          ],
        },
        custom_attributes: {
          attributes: [
            {
              id: "1",
              pointer: "/x_team",
              type: "string",
              access_control: {
                end_user: "hidden",
                bearer: "hidden",
                portal_ui: "readwrite",
              },
            },
          ],
        },
      },
    } as PortalAPIAppConfig;
    expect(knownPointers(config)).toEqual(["/gender", "/x_team"]);
  });
});
