# Auth Flows List and New Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** "Auth flows" opens on a list of the project's flows with create, open, set-as-default, rename, duplicate and delete, and a New flow dialog that starts a flow from scratch or a template; each flow (or Signup & login bundle) opens in the editor at its own URL.

**Architecture:** Three pure modules under `portal/src/graphql/portal/authflows/list/` (rows, actions, starters) carry every rule and are unit-tested. The list screen applies actions through `useAppConfigForm.saveOnly`, writing to the raw `authentication_flow`. The existing `AuthFlowsScreen` becomes the editor screen, reading the flow from the route and editing a bundle's three flows as one draft.

**Tech Stack:** React 18, React Router 6, Radix Themes, immer, Jest (`@jest/globals`), Apollo (existing `authFlowsScreenQuery`).

**Spec:** `docs/plans/auth-flow-builder/2026-10-19-01-flow-list-design.md` (DEV-3944). Builds on `2026-10-15-01-editor-design.md`.

## Global Constraints

- Branch `auth-flow-list`, stacked on `auth-flow-editor`. Commits end with `ref DEV-3944`.
- Flow names match `^[a-zA-Z_][a-zA-Z0-9_]*$`; unique within a type; a bundle's name unused in all three types.
- `default` is the only name AuthUI runs. Renaming to `default` is refused; use Set as default.
- Creatable types: `signup_login`, `signup`, `login`. Others are listed and editable only.
- List actions write to the raw `authentication_flow` (never the effective config, which carries defaulted `rate_limits`).
- All user-facing text through `locale-data/en.json`, static message IDs (portal `update-portal-ui` skill).
- Checks per task: `npx jest src/graphql/portal/authflows`, `npm run typecheck`, `npx eslint <touched>`, `npx stylelint` on touched CSS, run in `portal/`.

## Review Focus

- A signup_login flow whose options name flows of other names is not a bundle: its row has no "Sign up/Log in" subline and actions affect only it. Pinned in Task 1.
- Set as default when the old default is named by other signup_login flows: those references must follow the swap. Pinned in Task 2.
- Deleting a signup or login flow named by a signup_login flow outside the bundle being deleted is refused. Pinned in Task 2.
- A starter for a project without the needed login method is offered disabled, never built with an identification the project lacks. Pinned in Task 3.
- Opening `/flows/<type>/<name>` for a name that does not exist shows "Flow not found" with a link to the list, not a blank canvas. Pinned in Task 4.

---

### Task 1: List rows

**Files:**
- Create: `portal/src/graphql/portal/authflows/list/rows.ts`
- Test: `portal/src/graphql/portal/authflows/list/rows.test.ts`

**Interfaces:**
- Consumes: `FlowType`, `flowTypes`, `flowListKey`, `DEFAULT_FLOW_NAME` from `../model`.
- Produces:
  - `type CreatableType = "signup_login" | "signup" | "login"`, `CREATABLE_TYPES: readonly CreatableType[]`
  - `bundleTypes(): readonly FlowType[]` → `["signup_login", "signup", "login"]`
  - `isBundle(flows: AuthenticationFlowsConfig | undefined, name: string): boolean`
  - `interface FlowRow { type: FlowType; name: string; generated: boolean; inUse: boolean; bundle: boolean; partOf: string | null }`
  - `type SectionKey = "signup_login" | "signup" | "login" | "other"`
  - `interface FlowSection { key: SectionKey; rows: FlowRow[] }`
  - `flowSections(custom: AuthenticationFlowsConfig | undefined, generated: AuthenticationFlowsConfig): FlowSection[]`

Rules: a bundle is a signup_login flow `N` whose every option with `login_flow` names `N` and every option with `signup_flow` names `N`, with at least one hand-off, where `signup_flows` and `login_flows` both have `N` (custom, or the generated defaults when `N` is `default` and has no custom one). A type with no custom `default` gets a generated `default` row first. Rows otherwise keep config order. "Other" holds promote, reauth and account_recovery in that order, and is omitted when empty.

- [ ] **Step 1: Write the failing tests**

```ts
import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowsConfig } from "../../../../types";
import { flowSections, isBundle } from "./rows";

const generated: AuthenticationFlowsConfig = {
  signup_login_flows: [
    {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            { identification: "email", login_flow: "default", signup_flow: "default" },
          ],
        },
      ],
    },
  ],
  signup_flows: [{ name: "default", steps: [{ type: "identify" }] }],
  login_flows: [{ name: "default", steps: [{ type: "identify" }] }],
  account_recovery_flows: [{ name: "default", steps: [{ type: "identify" }] }],
};

const handoff = (name: string, login: string, signup: string) => ({
  name,
  steps: [
    {
      type: "identify",
      one_of: [
        { identification: "email", login_flow: login, signup_flow: signup },
        { identification: "passkey", login_flow: login },
      ],
    },
  ],
});

describe("flowSections", () => {
  it("lists the generated defaults when there are no custom flows", () => {
    const sections = flowSections(undefined, generated);
    expect(sections.map((s) => s.key)).toEqual([
      "signup_login",
      "signup",
      "login",
      "other",
    ]);
    expect(sections[0].rows).toEqual([
      { type: "signup_login", name: "default", generated: true, inUse: true, bundle: true, partOf: null },
    ]);
    expect(sections[1].rows[0]).toMatchObject({ generated: true, partOf: "default" });
    expect(sections[3].rows.map((r) => r.type)).toEqual(["account_recovery"]);
  });

  it("puts a custom default in place of the generated one", () => {
    const custom: AuthenticationFlowsConfig = {
      login_flows: [{ name: "default", steps: [] }, { name: "staff", steps: [] }],
    };
    const login = flowSections(custom, generated)[2].rows;
    expect(login.map((r) => [r.name, r.generated, r.inUse])).toEqual([
      ["default", false, true],
      ["staff", false, false],
    ]);
  });

  it("bundles a signup_login flow with same-named signup and login flows", () => {
    const custom: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("members", "members", "members")],
      signup_flows: [{ name: "members", steps: [] }],
      login_flows: [{ name: "members", steps: [] }],
    };
    const [sl, signup, login] = flowSections(custom, generated);
    expect(sl.rows.find((r) => r.name === "members")).toMatchObject({ bundle: true });
    expect(signup.rows.find((r) => r.name === "members")?.partOf).toEqual("members");
    expect(login.rows.find((r) => r.name === "members")?.partOf).toEqual("members");
  });

  it("does not bundle a signup_login flow that hands off to other names", () => {
    const custom: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("members", "default", "members")],
      signup_flows: [{ name: "members", steps: [] }],
      login_flows: [{ name: "members", steps: [] }],
    };
    expect(isBundle(custom, "members")).toEqual(false);
    const signup = flowSections(custom, generated)[1];
    expect(signup.rows.find((r) => r.name === "members")?.partOf).toEqual(null);
  });

  it("omits Other when the project has none", () => {
    const keys = flowSections(undefined, { ...generated, account_recovery_flows: [] }).map(
      (s) => s.key
    );
    expect(keys).toEqual(["signup_login", "signup", "login"]);
  });
});
```

- [ ] **Step 2: Run to verify they fail**

Run: `cd portal && npx jest src/graphql/portal/authflows/list/rows`
Expected: FAIL, cannot find module `./rows`.

- [ ] **Step 3: Implement**

```ts
import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";

export type CreatableType = "signup_login" | "signup" | "login";
export const CREATABLE_TYPES: readonly CreatableType[] = [
  "signup_login",
  "signup",
  "login",
];

const BUNDLE_TYPES: readonly FlowType[] = ["signup_login", "signup", "login"];
export function bundleTypes(): readonly FlowType[] {
  return BUNDLE_TYPES;
}

const OTHER_TYPES: readonly FlowType[] = ["promote", "reauth", "account_recovery"];

function flowsOf(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType
): AuthenticationFlowConfig[] {
  return config?.[flowListKey(type)] ?? [];
}

function has(
  config: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  name: string
): boolean {
  return flowsOf(config, type).some((f) => f.name === name);
}

// Whether the signup_login flow `name` only hands off to the signup and
// login flows of the same name, and both exist.
function bundles(
  flow: AuthenticationFlowConfig,
  hasMember: (type: FlowType) => boolean
): boolean {
  const options = (flow.steps ?? []).flatMap((s) => s.one_of ?? []);
  const handoffs = options.filter((o) => o.login_flow != null || o.signup_flow != null);
  return (
    handoffs.length > 0 &&
    options.every(
      (o) =>
        (o.login_flow == null || o.login_flow === flow.name) &&
        (o.signup_flow == null || o.signup_flow === flow.name)
    ) &&
    hasMember("signup") &&
    hasMember("login")
  );
}

export function isBundle(
  custom: AuthenticationFlowsConfig | undefined,
  name: string,
  generated?: AuthenticationFlowsConfig
): boolean {
  const lookup = (type: FlowType) =>
    flowsOf(custom, type).find((f) => f.name === name) ??
    (name === DEFAULT_FLOW_NAME ? flowsOf(generated, type).find((f) => f.name === name) : undefined);
  const flow = lookup("signup_login");
  return flow != null && bundles(flow, (t) => lookup(t) != null);
}

export interface FlowRow {
  type: FlowType;
  name: string;
  generated: boolean;
  inUse: boolean;
  bundle: boolean;
  partOf: string | null;
}

export type SectionKey = "signup_login" | "signup" | "login" | "other";

export interface FlowSection {
  key: SectionKey;
  rows: FlowRow[];
}

function rowsOf(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig,
  type: FlowType
): FlowRow[] {
  const names: { name: string; generated: boolean }[] = [];
  if (!has(custom, type, DEFAULT_FLOW_NAME) && has(generated, type, DEFAULT_FLOW_NAME)) {
    names.push({ name: DEFAULT_FLOW_NAME, generated: true });
  }
  for (const f of flowsOf(custom, type)) {
    names.push({ name: f.name, generated: false });
  }
  return names.map(({ name, generated: g }) => {
    const bundled = isBundle(custom, name, generated);
    return {
      type,
      name,
      generated: g,
      inUse: name === DEFAULT_FLOW_NAME,
      bundle: type === "signup_login" && bundled,
      partOf: (type === "signup" || type === "login") && bundled ? name : null,
    };
  });
}

export function flowSections(
  custom: AuthenticationFlowsConfig | undefined,
  generated: AuthenticationFlowsConfig
): FlowSection[] {
  const sections: FlowSection[] = (["signup_login", "signup", "login"] as const).map(
    (key) => ({ key, rows: rowsOf(custom, generated, key) })
  );
  const other = OTHER_TYPES.flatMap((t) => rowsOf(custom, generated, t));
  if (other.length > 0) {
    sections.push({ key: "other", rows: other });
  }
  return sections;
}
```

- [ ] **Step 4: Run to verify they pass**

Run: `cd portal && npx jest src/graphql/portal/authflows/list/rows`
Expected: PASS, 5 tests.

- [ ] **Step 5: Commit**

```bash
git add portal/src/graphql/portal/authflows/list/rows*
git commit -m "[Portal] Group auth flows into list rows" -m "ref DEV-3944"
```

### Task 2: List actions

**Files:**
- Create: `portal/src/graphql/portal/authflows/list/actions.ts`
- Test: `portal/src/graphql/portal/authflows/list/actions.test.ts`

**Interfaces:**
- Consumes: `bundleTypes`, `isBundle` from `./rows`; `flowListKey`, `FlowType`, `DEFAULT_FLOW_NAME` from `../model`.
- Produces:
  - `type FlowAction =`
    `| { kind: "set_default"; type: FlowType; name: string; bundle: boolean }`
    `| { kind: "rename"; type: FlowType; name: string; newName: string; bundle: boolean }`
    `| { kind: "add"; flows: { type: FlowType; flow: AuthenticationFlowConfig }[] }`
    `| { kind: "delete"; type: FlowType; name: string; bundle: boolean }`
  - `applyFlowAction(flows: AuthenticationFlowsConfig | undefined, action: FlowAction): AuthenticationFlowsConfig`
  - `copyAs(sources: { type: FlowType; flow: AuthenticationFlowConfig }[], newName: string): { type: FlowType; flow: AuthenticationFlowConfig }[]` (for Duplicate and templates: renames each copy and rewrites hand-offs from the old name to `newName`)
  - `deleteBlockers(flows: AuthenticationFlowsConfig | undefined, type: FlowType, name: string, bundle: boolean): string[]` (names of signup_login flows still handing off to it)
  - `type NameError = "invalid" | "taken" | "reserved"`
  - `nameError(flows: AuthenticationFlowsConfig | undefined, types: readonly FlowType[], name: string, allowDefault: boolean): NameError | null`

Duplicate is `add` with `copyAs(...)`; the dialog resolves generated sources itself.

- [ ] **Step 1: Write the failing tests**

```ts
import { describe, it, expect } from "@jest/globals";
import { AuthenticationFlowsConfig } from "../../../../types";
import { applyFlowAction, copyAs, deleteBlockers, nameError } from "./actions";

const handoff = (name: string, to: string) => ({
  name,
  steps: [
    {
      type: "identify",
      one_of: [{ identification: "email", login_flow: to, signup_flow: to }],
    },
  ],
});

const flows: AuthenticationFlowsConfig = {
  signup_login_flows: [handoff("default", "default"), handoff("members", "members")],
  signup_flows: [{ name: "default", steps: [] }, { name: "members", steps: [] }],
  login_flows: [{ name: "default", steps: [] }, { name: "members", steps: [] }],
};

const names = (f: AuthenticationFlowsConfig, key: keyof AuthenticationFlowsConfig) =>
  (f[key] ?? []).map((x) => x.name);
const targets = (f: AuthenticationFlowsConfig) =>
  (f.signup_login_flows ?? []).map((x) => [x.name, x.steps?.[0].one_of?.[0].login_flow]);

describe("set_default", () => {
  it("swaps a bundle's three flows with the defaults and their hand-offs", () => {
    const next = applyFlowAction(flows, { kind: "set_default", type: "signup_login", name: "members", bundle: true });
    expect(names(next, "signup_flows")).toEqual(["members", "default"]);
    expect(targets(next)).toEqual([["members", "members"], ["default", "default"]]);
  });

  it("makes a flow default when only the generated one existed", () => {
    const next = applyFlowAction({ login_flows: [{ name: "staff", steps: [] }] }, {
      kind: "set_default", type: "login", name: "staff", bundle: false,
    });
    expect(names(next, "login_flows")).toEqual(["default"]);
  });

  it("rewrites hand-offs in other signup_login flows that named either flow", () => {
    const f: AuthenticationFlowsConfig = {
      signup_login_flows: [handoff("other", "default")],
      login_flows: [{ name: "default", steps: [] }, { name: "staff", steps: [] }],
      signup_flows: [{ name: "default", steps: [] }],
    };
    const next = applyFlowAction(f, { kind: "set_default", type: "login", name: "staff", bundle: false });
    expect(next.signup_login_flows?.[0].steps?.[0].one_of?.[0]).toMatchObject({
      login_flow: "staff",
      signup_flow: "default",
    });
  });
});

describe("rename", () => {
  it("renames a bundle and its hand-offs", () => {
    const next = applyFlowAction(flows, {
      kind: "rename", type: "signup_login", name: "members", newName: "staff", bundle: true,
    });
    expect(names(next, "login_flows")).toEqual(["default", "staff"]);
    expect(targets(next)[1]).toEqual(["staff", "staff"]);
  });
});

describe("add and copyAs", () => {
  it("copies a bundle under a new name with its hand-offs", () => {
    const copies = copyAs(
      [
        { type: "signup_login", flow: flows.signup_login_flows![1] },
        { type: "signup", flow: flows.signup_flows![1] },
        { type: "login", flow: flows.login_flows![1] },
      ],
      "members_2"
    );
    const next = applyFlowAction(flows, { kind: "add", flows: copies });
    expect(names(next, "signup_flows")).toEqual(["default", "members", "members_2"]);
    expect(targets(next)[2]).toEqual(["members_2", "members_2"]);
  });
});

describe("delete", () => {
  it("deletes a bundle's three flows", () => {
    const next = applyFlowAction(flows, { kind: "delete", type: "signup_login", name: "members", bundle: true });
    expect(names(next, "login_flows")).toEqual(["default"]);
  });

  it("blocks deleting a login flow a signup_login flow still names", () => {
    expect(deleteBlockers(flows, "login", "members", false)).toEqual(["members"]);
    expect(deleteBlockers(flows, "signup_login", "members", true)).toEqual([]);
  });
});

describe("nameError", () => {
  it("checks the pattern, clashes across a bundle's types, and default", () => {
    const all = ["signup_login", "signup", "login"] as const;
    expect(nameError(flows, all, "2fa", false)).toEqual("invalid");
    expect(nameError(flows, all, "members", false)).toEqual("taken");
    expect(nameError(flows, all, "default", false)).toEqual("taken");
    expect(nameError({}, ["login"], "default", false)).toEqual("reserved");
    expect(nameError({}, ["login"], "default", true)).toEqual(null);
    expect(nameError(flows, all, "staff", false)).toEqual(null);
  });
});
```

- [ ] **Step 2: Run to verify they fail**

Run: `cd portal && npx jest src/graphql/portal/authflows/list/actions`
Expected: FAIL, cannot find module `./actions`.

- [ ] **Step 3: Implement**

```ts
import { produce } from "immer";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";
import { bundleTypes } from "./rows";

export type FlowAction =
  | { kind: "set_default"; type: FlowType; name: string; bundle: boolean }
  | { kind: "rename"; type: FlowType; name: string; newName: string; bundle: boolean }
  | { kind: "add"; flows: { type: FlowType; flow: AuthenticationFlowConfig }[] }
  | { kind: "delete"; type: FlowType; name: string; bundle: boolean };

const NAME_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

function typesOf(type: FlowType, bundle: boolean): readonly FlowType[] {
  return bundle ? bundleTypes() : [type];
}

// Renames per type, as old name → new name, then points every signup_login
// hand-off at the new names.
function renameAll(
  d: AuthenticationFlowsConfig,
  renames: { type: FlowType; from: string; to: string }[]
) {
  for (const { type, from, to } of renames) {
    for (const f of d[flowListKey(type)] ?? []) {
      if (f.name === from) {
        f.name = to;
      }
    }
  }
  const map = (type: "login" | "signup", v: string | undefined) => {
    const r = renames.find((x) => x.type === type && x.from === v);
    return r != null ? r.to : v;
  };
  for (const f of d.signup_login_flows ?? []) {
    for (const step of f.steps ?? []) {
      for (const o of step.one_of ?? []) {
        if (o.login_flow != null) o.login_flow = map("login", o.login_flow);
        if (o.signup_flow != null) o.signup_flow = map("signup", o.signup_flow);
      }
    }
  }
}

export function copyAs(
  sources: { type: FlowType; flow: AuthenticationFlowConfig }[],
  newName: string
): { type: FlowType; flow: AuthenticationFlowConfig }[] {
  return sources.map(({ type, flow }) => {
    const copy = produce(flow, (d) => {
      const old = d.name;
      d.name = newName;
      for (const step of d.steps ?? []) {
        for (const o of step.one_of ?? []) {
          if (o.login_flow === old) o.login_flow = newName;
          if (o.signup_flow === old) o.signup_flow = newName;
        }
      }
    });
    return { type, flow: copy };
  });
}

export function applyFlowAction(
  flows: AuthenticationFlowsConfig | undefined,
  action: FlowAction
): AuthenticationFlowsConfig {
  return produce(flows ?? {}, (d) => {
    switch (action.kind) {
      case "add":
        for (const { type, flow } of action.flows) {
          (d[flowListKey(type)] ??= []).push(flow);
        }
        return;
      case "delete":
        for (const t of typesOf(action.type, action.bundle)) {
          d[flowListKey(t)] = (d[flowListKey(t)] ?? []).filter((f) => f.name !== action.name);
        }
        return;
      case "rename":
        renameAll(
          d,
          typesOf(action.type, action.bundle).map((t) => ({ type: t, from: action.name, to: action.newName }))
        );
        return;
      case "set_default": {
        // A placeholder no flow can be named keeps the swap from colliding.
        const tmp = " swapping";
        const types = typesOf(action.type, action.bundle);
        renameAll(d, types.map((t) => ({ type: t, from: DEFAULT_FLOW_NAME, to: tmp })));
        renameAll(d, types.map((t) => ({ type: t, from: action.name, to: DEFAULT_FLOW_NAME })));
        renameAll(d, types.map((t) => ({ type: t, from: tmp, to: action.name })));
      }
    }
  });
}

export function deleteBlockers(
  flows: AuthenticationFlowsConfig | undefined,
  type: FlowType,
  name: string,
  bundle: boolean
): string[] {
  if (type !== "signup" && type !== "login" && !bundle) {
    return [];
  }
  const key = type === "login" ? "login_flow" : "signup_flow";
  return (flows?.signup_login_flows ?? [])
    .filter((f) => !(bundle && f.name === name))
    .filter((f) =>
      (f.steps ?? []).some((s) =>
        (s.one_of ?? []).some((o) =>
          bundle ? o.login_flow === name || o.signup_flow === name : o[key] === name
        )
      )
    )
    .map((f) => f.name);
}

export type NameError = "invalid" | "taken" | "reserved";

export function nameError(
  flows: AuthenticationFlowsConfig | undefined,
  types: readonly FlowType[],
  name: string,
  allowDefault: boolean
): NameError | null {
  if (!NAME_PATTERN.test(name)) {
    return "invalid";
  }
  if (types.some((t) => (flows?.[flowListKey(t)] ?? []).some((f) => f.name === name))) {
    return "taken";
  }
  if (name === DEFAULT_FLOW_NAME && !allowDefault) {
    return "reserved";
  }
  return null;
}
```

- [ ] **Step 4: Run to verify they pass**

Run: `cd portal && npx jest src/graphql/portal/authflows/list/actions`
Expected: PASS, 8 tests.

- [ ] **Step 5: Commit**

```bash
git add portal/src/graphql/portal/authflows/list/actions*
git commit -m "[Portal] Add set-as-default, rename, duplicate and delete for auth flows" -m "ref DEV-3944"
```

### Task 3: Starters (from scratch and templates)

**Files:**
- Create: `portal/src/graphql/portal/authflows/list/starters.ts`
- Test: `portal/src/graphql/portal/authflows/list/starters.test.ts`

**Interfaces:**
- Consumes: `CreatableType` from `./rows`; `copyAs` from `./actions`; `FlowType`, `flowListKey`, `DEFAULT_FLOW_NAME` from `../model`; `allowedStepTypes` from `../edit/rules`.
- Produces:
  - `type NewFlows = { type: FlowType; flow: AuthenticationFlowConfig }[]`
  - `interface ProjectMethods { identifications: string[]; primary: string[]; secondary: string[] }` with `projectMethods(config: PortalAPIAppConfig): ProjectMethods` (login ID key types plus `oauth`/`passkey` identities; `primary_*` / `secondary_*` authentication values the project enables)
  - `scratchFlows(type: CreatableType, name: string, methods: ProjectMethods): NewFlows`
  - `type StarterID = "current" | "email_password" | "email_code" | "phone_code" | "email_password_2fa"`
  - `STARTER_IDS: readonly StarterID[]`
  - `type MissingMethod = "email_login" | "phone_login" | "password" | "email_code" | "sms_code" | "totp"`
  - `starterMissing(id: StarterID, methods: ProjectMethods): MissingMethod[]` (empty when available)
  - `starterFlows(id: StarterID, type: CreatableType, name: string, generated: AuthenticationFlowsConfig): NewFlows`
  - `stepSummary(flow: AuthenticationFlowConfig): string[]` (step types along the first option of each choice)

Starters are written as a signup flow and a login flow; Signup & login adds a signup_login flow handing off to both. "current" copies the generated `default` flows of the type (all three for Signup & login) with `copyAs`.

- [ ] **Step 1: Write the failing tests**

```ts
import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import { allowedStepTypes } from "../edit/rules";
import {
  STARTER_IDS,
  projectMethods,
  scratchFlows,
  starterFlows,
  starterMissing,
  stepSummary,
} from "./starters";

const config: PortalAPIAppConfig = {
  identity: { login_id: { keys: [{ type: "email" }, { type: "username" }] } },
  authentication: {
    identities: ["login_id", "oauth"],
    primary_authenticators: ["password", "oob_otp_email"],
    secondary_authenticators: ["totp"],
  },
} as PortalAPIAppConfig;

const generated = {
  signup_login_flows: [{ name: "default", steps: [{ type: "identify", one_of: [{ identification: "email", login_flow: "default", signup_flow: "default" }] }] }],
  signup_flows: [{ name: "default", steps: [{ type: "identify", one_of: [{ identification: "email" }] }] }],
  login_flows: [{ name: "default", steps: [{ type: "identify", one_of: [{ identification: "email" }] }] }],
};

describe("projectMethods", () => {
  it("reads login IDs, identities and authenticators", () => {
    expect(projectMethods(config)).toEqual({
      identifications: ["email", "username", "oauth"],
      primary: ["primary_password", "primary_oob_otp_email"],
      secondary: ["secondary_totp"],
    });
  });
});

describe("scratchFlows", () => {
  it("builds a bundle handing off to its own signup and login flows", () => {
    const out = scratchFlows("signup_login", "members", projectMethods(config));
    expect(out.map((f) => [f.type, f.flow.name])).toEqual([
      ["signup_login", "members"],
      ["signup", "members"],
      ["login", "members"],
    ]);
    expect(out[0].flow.steps?.[0].one_of?.[0]).toEqual({
      identification: "email",
      login_flow: "members",
      signup_flow: "members",
    });
    expect(out[2].flow.steps?.map((s) => s.type)).toEqual(["identify", "authenticate"]);
  });
});

describe("starterMissing", () => {
  it("names what the project lacks", () => {
    const m = projectMethods(config);
    expect(starterMissing("email_password", m)).toEqual([]);
    expect(starterMissing("phone_code", m)).toEqual(["phone_login", "sms_code"]);
    expect(starterMissing("current", m)).toEqual([]);
  });
});

describe("starterFlows", () => {
  it("copies the generated flows under the new name", () => {
    const out = starterFlows("current", "signup_login", "members", generated);
    expect(out[0].flow.steps?.[0].one_of?.[0]).toMatchObject({ login_flow: "members", signup_flow: "members" });
  });

  it("uses only step types the editor allows", () => {
    for (const id of STARTER_IDS.filter((x) => x !== "current")) {
      for (const type of ["signup_login", "signup", "login"] as const) {
        for (const { type: t, flow } of starterFlows(id, type, "x", generated)) {
          const walk = (steps: typeof flow.steps): string[] =>
            (steps ?? []).flatMap((s) => [s.type, ...(s.one_of ?? []).flatMap((o) => walk(o.steps))]);
          for (const stepType of walk(flow.steps)) {
            expect(allowedStepTypes(t)).toContain(stepType);
          }
        }
      }
    }
  });
});

describe("stepSummary", () => {
  it("follows the first option of each choice", () => {
    const [signup] = starterFlows("email_password", "signup", "x", generated);
    expect(stepSummary(signup.flow)).toEqual(["identify", "verify", "create_authenticator"]);
  });
});
```

- [ ] **Step 2: Run to verify they fail**

Run: `cd portal && npx jest src/graphql/portal/authflows/list/starters`
Expected: FAIL, cannot find module `./starters`.

- [ ] **Step 3: Implement**

```ts
import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";
import { CreatableType } from "./rows";
import { copyAs } from "./actions";

export type NewFlows = { type: FlowType; flow: AuthenticationFlowConfig }[];

export interface ProjectMethods {
  identifications: string[];
  primary: string[];
  secondary: string[];
}

export function projectMethods(config: PortalAPIAppConfig): ProjectMethods {
  const identities = config.authentication?.identities ?? [];
  const loginIDs = identities.includes("login_id")
    ? (config.identity?.login_id?.keys ?? []).map((k) => k.type)
    : [];
  return {
    identifications: [
      ...loginIDs,
      ...identities.filter((i) => i === "oauth" || i === "passkey"),
    ],
    primary: (config.authentication?.primary_authenticators ?? []).map((a) => `primary_${a}`),
    secondary: (config.authentication?.secondary_authenticators ?? []).map((a) => `secondary_${a}`),
  };
}

// Identifications that can lead to signup; passkey and oauth only log in here.
const LOGIN_ONLY = new Set(["passkey"]);

function bundle(name: string, identification: string, signup: AuthenticationFlowConfig, login: AuthenticationFlowConfig): NewFlows {
  return [
    {
      type: "signup_login",
      flow: {
        name,
        steps: [
          {
            type: "identify",
            one_of: [
              LOGIN_ONLY.has(identification)
                ? { identification, login_flow: name }
                : { identification, login_flow: name, signup_flow: name },
            ],
          },
        ],
      },
    },
    { type: "signup", flow: signup },
    { type: "login", flow: login },
  ];
}

export function scratchFlows(type: CreatableType, name: string, methods: ProjectMethods): NewFlows {
  const identification = methods.identifications[0] ?? "email";
  const authentication = methods.primary[0] ?? "primary_password";
  const signup: AuthenticationFlowConfig = {
    name,
    steps: [{ type: "identify", one_of: [{ identification }] }],
  };
  const login: AuthenticationFlowConfig = {
    name,
    steps: [
      { type: "identify", one_of: [{ identification }] },
      { type: "authenticate", one_of: [{ authentication }] },
    ],
  };
  if (type === "signup") return [{ type, flow: signup }];
  if (type === "login") return [{ type, flow: login }];
  return bundle(name, identification, signup, login);
}

export type StarterID = "current" | "email_password" | "email_code" | "phone_code" | "email_password_2fa";
export const STARTER_IDS: readonly StarterID[] = [
  "current",
  "email_password",
  "email_code",
  "phone_code",
  "email_password_2fa",
];

export type MissingMethod = "email_login" | "phone_login" | "password" | "email_code" | "sms_code" | "totp";

const needs: Record<StarterID, { identification?: string; primary?: string; secondary?: string }> = {
  current: {},
  email_password: { identification: "email", primary: "primary_password" },
  email_code: { identification: "email", primary: "primary_oob_otp_email" },
  phone_code: { identification: "phone", primary: "primary_oob_otp_sms" },
  email_password_2fa: { identification: "email", primary: "primary_password", secondary: "secondary_totp" },
};

const missingOf: Record<string, MissingMethod> = {
  email: "email_login",
  phone: "phone_login",
  primary_password: "password",
  primary_oob_otp_email: "email_code",
  primary_oob_otp_sms: "sms_code",
  secondary_totp: "totp",
};

export function starterMissing(id: StarterID, methods: ProjectMethods): MissingMethod[] {
  const n = needs[id];
  const out: MissingMethod[] = [];
  if (n.identification != null && !methods.identifications.includes(n.identification)) out.push(missingOf[n.identification]);
  if (n.primary != null && !methods.primary.includes(n.primary)) out.push(missingOf[n.primary]);
  if (n.secondary != null && !methods.secondary.includes(n.secondary)) out.push(missingOf[n.secondary]);
  return out;
}

// The signup and login flows of a starter.
function starterPair(id: Exclude<StarterID, "current">, name: string): { identification: string; signup: AuthenticationFlowConfig; login: AuthenticationFlowConfig } {
  const identify = (identification: string, steps: AuthenticationFlowConfig["steps"]) => ({
    type: "identify",
    name: "identify",
    one_of: [{ identification, steps }],
  });
  switch (id) {
    case "email_password":
    case "email_password_2fa": {
      const twoFA = id === "email_password_2fa";
      return {
        identification: "email",
        signup: {
          name,
          steps: [
            identify("email", [
              { type: "verify", target_step: "identify" },
              { type: "create_authenticator", one_of: [{ authentication: "primary_password" }] },
              ...(twoFA
                ? [
                    { type: "create_authenticator", one_of: [{ authentication: "secondary_totp" }] },
                    { type: "view_recovery_code" },
                  ]
                : []),
            ]),
          ],
        },
        login: {
          name,
          steps: [
            identify("email", [
              { type: "authenticate", one_of: [{ authentication: "primary_password" }] },
              ...(twoFA
                ? [{ type: "authenticate", one_of: [{ authentication: "secondary_totp" }, { authentication: "recovery_code" }] }]
                : []),
            ]),
          ],
        },
      };
    }
    case "email_code":
    case "phone_code": {
      const identification = id === "email_code" ? "email" : "phone";
      const authentication = id === "email_code" ? "primary_oob_otp_email" : "primary_oob_otp_sms";
      const codeStep = (type: string) => ({ type, one_of: [{ authentication, target_step: "identify" }] });
      return {
        identification,
        signup: { name, steps: [identify(identification, [codeStep("create_authenticator")])] },
        login: { name, steps: [identify(identification, [codeStep("authenticate")])] },
      };
    }
  }
}

export function starterFlows(id: StarterID, type: CreatableType, name: string, generated: AuthenticationFlowsConfig): NewFlows {
  if (id === "current") {
    const types: FlowType[] = type === "signup_login" ? ["signup_login", "signup", "login"] : [type];
    const sources = types.flatMap((t) =>
      (generated[flowListKey(t)] ?? []).filter((f) => f.name === DEFAULT_FLOW_NAME).map((flow) => ({ type: t, flow }))
    );
    return copyAs(sources, name);
  }
  const { identification, signup, login } = starterPair(id, name);
  if (type === "signup") return [{ type, flow: signup }];
  if (type === "login") return [{ type, flow: login }];
  return bundle(name, identification, signup, login);
}

export function stepSummary(flow: AuthenticationFlowConfig): string[] {
  const walk = (steps: AuthenticationFlowConfig["steps"]): string[] =>
    (steps ?? []).flatMap((s) => [s.type, ...walk(s.one_of?.[0]?.steps)]);
  return walk(flow.steps);
}
```

- [ ] **Step 4: Run to verify they pass**

Run: `cd portal && npx jest src/graphql/portal/authflows/list/starters`
Expected: PASS, 6 tests. If `allowedStepTypes` rejects a starter step (for example `view_recovery_code` in signup), the test names it: fix the starter, not the test.

- [ ] **Step 5: Commit**

```bash
git add portal/src/graphql/portal/authflows/list/starters*
git commit -m "[Portal] Add from-scratch and template auth flows" -m "ref DEV-3944"
```

### Task 4: Editor at its own URL, with bundle tabs

**Files:**
- Modify: `portal/src/graphql/portal/AuthFlowsScreen.tsx` (becomes the editor; keep the file and default export so the lazy import stays)
- Modify: `portal/src/graphql/portal/authflows/edit/save.ts` (multi-flow draft)
- Modify: `portal/src/graphql/portal/authflows/edit/save.test.ts`
- Modify: `portal/src/graphql/portal/authflows/FlowSketchContext.tsx`, `FlowNodes.tsx` (hand-off links)
- Modify: `portal/src/graphql/portal/authflows/model.ts` (drop `parseFlowSelection`, `flowSelectionSearch`, `listFlowNames` only if unused; add `flowPath`)
- Modify: `portal/src/AppRoot.tsx` (route `flows/:type/:name`)
- Modify: `portal/src/locale-data/en.json`

**Interfaces:**
- Consumes: `isBundle`, `bundleTypes` from `./authflows/list/rows`; `NewFlows` from `./authflows/list/starters`.
- Produces:
  - `EditorState = { drafts: EditorDraft[] }` (empty while viewing); `replaceFlows(raw, drafts): PortalAPIAppConfig`
  - `flowPath(appID: string, type: FlowType, name: string): string` → `/project/${appID}/configuration/authentication/flows/${type}/${name}`
  - Router state for a new flow: `{ newFlows: NewFlows }`
  - `FlowSketchContextValue.flowLink(ref: FlowRef): string` used by the hand-off card's Open

Behaviour:
- `:type` must be a `FlowType`; otherwise, or when neither a custom nor (for `default`) a generated flow exists and no `newFlows` came in, show "Flow not found" with a link back to the list.
- A bundle (`isBundle` on the signup_login flow, custom or generated default) shows tabs Choose / Sign up / Log in (`?part=` absent, `signup`, `login`). The canvas shows the part's flow. Edit starts a draft of all three flows; Save writes all three with `replaceFlows`.
- Hand-off Open: inside a bundle, the matching tab; otherwise `flowPath` of the target.
- Header: back link "Auth flows" (`…/flows`), name and type, In use / Generated badges, Edit / Discard / Save. No pickers.
- New flow: when `location.state.newFlows` is set and the form has no drafts, set `drafts` from it once. Discard when none of the drafts exists in `rawConfig` navigates to the list.

- [ ] **Step 1: Write the failing test for `replaceFlows`**

```ts
it("replaces or appends every draft", () => {
  const raw = { authentication_flow: { login_flows: [{ name: "a", steps: [] }] } } as PortalAPIAppConfig;
  const next = replaceFlows(raw, [
    { type: "login", name: "a", flow: { name: "a", steps: [{ type: "identify" }] } },
    { type: "signup", name: "a", flow: { name: "a", steps: [] } },
  ]);
  expect(next.authentication_flow?.login_flows?.[0].steps).toHaveLength(1);
  expect(next.authentication_flow?.signup_flows?.map((f) => f.name)).toEqual(["a"]);
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd portal && npx jest src/graphql/portal/authflows/edit/save`
Expected: FAIL, `replaceFlows` is not exported.

- [ ] **Step 3: Implement `replaceFlows` and `EditorState.drafts`**

```ts
export interface EditorState {
  drafts: EditorDraft[];
}

export function replaceFlows(raw: PortalAPIAppConfig, drafts: EditorDraft[]): PortalAPIAppConfig {
  return drafts.reduce((r, d) => replaceFlow(r, d.type, d.name, d.flow), raw);
}
```

- [ ] **Step 4: Rework the screen**

In `AuthFlowsScreen.tsx`:
- `constructFormState` returns `{ drafts: [] }`; `constructConfig` returns `replaceFlows(raw, current.drafts)`.
- Read `type`, `name` from `useParams()` and `part` from `useSearchParams()`. `bundle = type === "signup_login" && isBundle(custom ?? undefined, name, generated)`. The shown flow's type is `part === "signup" ? "signup" : part === "login" ? "login" : type` when `bundle`.
- `draft` is `drafts.find((d) => d.type === shownType && d.name === name) ?? null`; `apply` updates that draft in `drafts`.
- `startEditing` builds drafts for `bundle ? bundleTypes() : [type]`, each from `rawFlow` or the resolved (generated) flow named `name`.
- Remove the type and name `Select`s, `onTypeChange`, `onNameChange`, `select`. Add the back link (`Link` from `../../Link`) and, when `bundle`, a Radix `SegmentedControl` with the three parts that sets `?part=`.
- Pass `flowLink` in the sketch context: `(ref) => bundle && ref.name === name ? `?part=${ref.type}` : flowPath(appID, ref.type, ref.name)`.
- `useEffect` on mount: `const incoming = (location.state as { newFlows?: NewFlows } | null)?.newFlows; if (incoming != null && form.state.drafts.length === 0) setState(() => ({ drafts: incoming.map(({ type: t, flow }) => ({ type: t, name: flow.name, flow })) }))`.
- `onDiscard`: `form.reset()`; when no draft exists in `form.rawConfig`, `navigate(listPath)`.
- Not found: `!isFlowType(type)`, or no draft, no custom flow and no generated default for the shown part.

In `FlowNodes.tsx` `FlowRefNode`, replace `to={{ search: flowSelectionSearch(ref) }}` with `to={flowLink(ref)}` from `useFlowSketchContext()`.

In `AppRoot.tsx`, add next to `path="flows"`:

```tsx
<Route
  path="flows/:type/:name"
  element={
    <Suspense fallback={<ShowLoading />}>
      <AuthFlowsScreen />
    </Suspense>
  }
/>
```

and point `path="flows"` at `AuthFlowsListScreen` in Task 5 (until then it keeps the editor so nothing breaks).

Strings: `AuthFlowsScreen.back` "Auth flows", `AuthFlowsScreen.part.choose` "Choose", `AuthFlowsScreen.part.signup` "Sign up", `AuthFlowsScreen.part.login` "Log in", `AuthFlowsScreen.badge.in-use` "In use". Remove `AuthFlowsScreen.picker.type`, `AuthFlowsScreen.picker.name` once unused.

- [ ] **Step 5: Run the checks**

Run: `cd portal && npx jest src/graphql/portal/authflows && npm run typecheck && npx eslint src/graphql/portal/AuthFlowsScreen.tsx src/graphql/portal/authflows src/AppRoot.tsx`
Expected: all pass, 0 errors.

- [ ] **Step 6: Commit**

```bash
git add portal/src
git commit -m "[Portal] Open each auth flow at its own URL, with tabs for a bundle" -m "ref DEV-3944"
```

### Task 5: The list screen and its actions

**Files:**
- Create: `portal/src/graphql/portal/AuthFlowsListScreen.tsx`, `AuthFlowsListScreen.module.css`
- Create: `portal/src/graphql/portal/authflows/list/FlowNameDialog.tsx` (rename and duplicate)
- Modify: `portal/src/AppRoot.tsx` (`path="flows"` → list)
- Modify: `portal/src/locale-data/en.json`

**Interfaces:**
- Consumes: `flowSections`, `FlowRow` (Task 1); `applyFlowAction`, `copyAs`, `deleteBlockers`, `nameError`, `FlowAction` (Task 2); `flowPath` (Task 4); `resolveFlow` from `./authflows/model`.
- Produces: `AuthFlowsListScreen` (default export); `FlowNameDialog({ open, title, types, initialName, allowDefault, flows, onSubmit(name), onCancel })`.

Behaviour:
- `useAppConfigForm<{ action: FlowAction | null }>`: `constructFormState` → `{ action: null }`; `constructConfig(raw, _i, current)` → `current.action == null ? raw : produce(raw, (d) => { d.authentication_flow = applyFlowAction(raw.authentication_flow, current.action) })`. Each action calls `form.saveOnly(() => ({ action }))`.
- Generated flows from `useAuthFlowsScreenQueryQuery` (`fetchPolicy: "network-only"`).
- Page: `Heading` "Auth flows", description, **New flow** button (opens Task 6's dialog), then one section per `flowSections` entry with a Radix `Table` (Name, Status, ⋯). Row name links to `flowPath`. Bundle rows add "Sign up: N · Log in: N"; rows with `partOf` show "Part of N"; generated rows a "Generated" badge.
- ⋯ menu (Radix `DropdownMenu`): Open; Set as default (hidden when `inUse` or `generated`); Rename (hidden for `generated`); Duplicate; Delete (hidden for `generated`; disabled with the blockers as tooltip when `deleteBlockers` is non-empty).
- Set as default and Delete confirm with `ConfirmationDialog`. Delete of an in-use flow says the generated flow takes over.
- Duplicate of a generated row copies the generated `default` flows (`resolveFlow` per bundle type) with `copyAs`.
- Name errors map to `AuthFlowsScreen.list.name.invalid` / `.taken` / `.reserved`.
- `FormErrorMessageBar` shows save errors.

Strings: `AuthFlowsScreen.list.*` for the title, description, section titles (`signup_login`, `signup`, `login`, `other`), column headers, actions, dialogs, badges ("Not in use", "Part of {name}", "Sign up: {name} · Log in: {name}").

- [ ] **Step 1: Build the screen and dialog as above; route `flows` to it.**
- [ ] **Step 2: Run the checks**

Run: `cd portal && npx jest src/graphql/portal/authflows && npm run typecheck && npx eslint src/graphql/portal src/AppRoot.tsx && npx stylelint "src/graphql/portal/**/*.css"`
Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add portal/src
git commit -m "[Portal] List auth flows with set as default, rename, duplicate and delete" -m "ref DEV-3944"
```

### Task 6: New flow dialog

**Files:**
- Create: `portal/src/graphql/portal/authflows/list/NewFlowDialog.tsx`, `NewFlowDialog.module.css`
- Modify: `portal/src/graphql/portal/AuthFlowsListScreen.tsx`
- Modify: `portal/src/locale-data/en.json`

**Interfaces:**
- Consumes: `CREATABLE_TYPES` (Task 1); `nameError` (Task 2); `projectMethods`, `scratchFlows`, `starterFlows`, `starterMissing`, `stepSummary`, `STARTER_IDS`, `NewFlows` (Task 3); `flowPath` (Task 4); `stepTitle` from `../labels`.
- Produces: `NewFlowDialog({ open, onOpenChange, custom, generated, config, appID })`; on Create, `navigate(flowPath(appID, type, name), { state: { newFlows } })`.

Steps of the dialog (one Radix `Dialog`, a `step` state):
1. Type cards for `CREATABLE_TYPES`, each with a title and sentence.
2. Start: From scratch; Use a template; Import (disabled, "Coming soon").
3. Template (only for "Use a template"): `STARTER_IDS` as cards. Each shows `stepSummary` of its first flow (signup flow for signup and bundles, login flow for login) mapped through `stepTitle` and joined with " → ". Disabled with `starterMissing` reasons and a `Link` to `…/configuration/authentication/login-methods` when unavailable.
4. Name: a `TextField` prefilled with the first free `${type}_${n}` (n from 2), errors from `nameError(custom, types, name, allowDefault)` where `types` is the bundle's three for Signup & login and `allowDefault` is true when no custom `default` exists for those types. With `default`, a note says the flow is in use once saved.

Back and Next buttons move between steps; Create on the last.

Strings: `AuthFlowsScreen.new.*` for every label, the type sentences, starter names, missing-method reasons ("Needs email login", "Needs phone login", "Needs password", "Needs email codes", "Needs SMS codes", "Needs an authenticator app"), and "Coming soon".

- [ ] **Step 1: Build the dialog; open it from the list's New flow button.**
- [ ] **Step 2: Run the checks** (same command as Task 5). Expected: all pass.
- [ ] **Step 3: Commit**

```bash
git add portal/src
git commit -m "[Portal] Create auth flows from scratch or a template" -m "ref DEV-3944"
```

### Task 7: Finish

- [ ] Remove now-unused exports (`parseFlowSelection`, `flowSelectionSearch`, `listFlowNames` if unused) and orphaned strings (diff the key set against `auth-flow-editor`).
- [ ] Update `2026-10-19-01-flow-list-design.md` for any deviation ruled during execution.
- [ ] Run: `cd portal && npx jest src/graphql/portal/authflows && npm run typecheck && npx eslint src && npx stylelint "src/**/*.css"`. Expected: all pass.
- [ ] Commit `chore: Remove the auth flow pickers' leftovers` with `ref DEV-3944`.
