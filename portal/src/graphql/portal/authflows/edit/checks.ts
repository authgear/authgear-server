import {
  AuthenticationFlowConfig,
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { FlowType } from "../model";
import { InsertionPoint, getAt, parsePointer } from "./pointer";

// Problems the editor finds itself, before the server sees the flow. Forms
// follow the server's checks (pkg/lib/config/authentication_flow_form.go,
// on the form branch), which reject the config; the rest only warn.

type Step = AuthenticationFlowStepConfig;
type Option = AuthenticationFlowOptionConfig;
type Form = Step["form"];

export type StepCheck =
  // Rejected on save.
  | { kind: "form_before_identify" }
  // option: the identify option whose form it is; absent on a form step.
  | { kind: "field_not_fillable"; pointer: string; option?: number }
  | { kind: "field_duplicate"; pointer: string; option?: number }
  // Saved, but the form shows before the user has signed in.
  | { kind: "form_before_authentication" };

export type FlowCheck = "no_check_account_status";

const formFlowTypes = new Set<FlowType>(["signup", "promote", "login"]);

// End users never fill these in a form: they are identities, or one part
// of the address, which is asked for whole.
export function neverFillable(pointer: string): boolean {
  return (
    pointer === "/email" ||
    pointer === "/phone_number" ||
    pointer === "/preferred_username" ||
    pointer.startsWith("/address/")
  );
}

// Attributes a form can ask for: those the end user can read and write.
export function fillablePointers(config: PortalAPIAppConfig): string[] {
  const standard = (
    config.user_profile?.standard_attributes?.access_control ?? []
  )
    .filter(
      (a) =>
        a.access_control.end_user === "readwrite" && !neverFillable(a.pointer)
    )
    .map((a) => a.pointer);
  const custom = (config.user_profile?.custom_attributes?.attributes ?? [])
    .filter((a) => a.access_control.end_user === "readwrite")
    .map((a) => a.pointer);
  return [...standard, ...custom];
}

// Every attribute of the project, whether a form can ask for it or not.
export function knownPointers(config: PortalAPIAppConfig): string[] {
  return [
    ...(config.user_profile?.standard_attributes?.access_control ?? []).map(
      (a) => a.pointer
    ),
    ...(config.user_profile?.custom_attributes?.attributes ?? []).map(
      (a) => a.pointer
    ),
  ];
}

// Each level of the path to pointer: the steps list, the step's index in it,
// and for every level but the last, the option taken from that step.
interface Level {
  steps: Step[];
  index: number;
  option: number | null;
}

// pointer may name a step, or where one would be inserted.
function levels(flow: AuthenticationFlowConfig, pointer: string): Level[] {
  const tokens = parsePointer(pointer);
  const out: Level[] = [];
  let stepsPointer = "";
  // tokens: steps, i, (one_of, j, steps, i)*
  for (let t = 0; t + 1 < tokens.length; t += 4) {
    stepsPointer = `${stepsPointer}/${tokens[t]}`;
    const index = Number(tokens[t + 1]);
    const option = t + 3 < tokens.length ? Number(tokens[t + 3]) : null;
    out.push({
      steps: (getAt(flow, stepsPointer) as Step[] | undefined) ?? [],
      index,
      option,
    });
    stepsPointer = `${stepsPointer}/${index}/one_of/${option ?? 0}`;
  }
  return out;
}

// Whether an identify step has run by the time the step at pointer runs, the
// way the server decides it for forms: an identify step earlier in the same
// list or an outer one, or the identify step whose option the step is under.
export function identifiedBefore(
  flow: AuthenticationFlowConfig,
  pointer: string
): boolean {
  return levels(flow, pointer).some(
    ({ steps, index, option }) =>
      steps.slice(0, index).some((s) => s.type === "identify") ||
      (option != null && steps[index]?.type === "identify")
  );
}

// Identifications that sign the user in by themselves; select_account
// continues a session the user already signed in to.
const AUTHENTICATING_IDENTIFICATIONS = new Set([
  "oauth",
  "passkey",
  "ldap",
  "select_account",
]);

function optionAuthenticates(o: Option): boolean {
  return (
    AUTHENTICATING_IDENTIFICATIONS.has(o.identification ?? "") ||
    (o.steps ?? []).some(stepAuthenticates)
  );
}

// Whether every way through the step signs the user in.
function stepAuthenticates(step: Step): boolean {
  if (step.type === "authenticate") {
    return true;
  }
  if (step.type === "identify") {
    const options = step.one_of ?? [];
    return options.length > 0 && options.every(optionAuthenticates);
  }
  return false;
}

// Whether the user has signed in by the time the step at pointer runs, on
// the path it is on: an earlier step that signs in on every way through it,
// an authenticate step it is under, or an option that signs in by itself.
export function authenticatedBefore(
  flow: AuthenticationFlowConfig,
  pointer: string
): boolean {
  return levels(flow, pointer).some(({ steps, index, option }) => {
    if (steps.slice(0, index).some(stepAuthenticates)) {
      return true;
    }
    if (option == null) {
      return false;
    }
    const parent = steps[index] as Step | undefined;
    if (parent?.type === "authenticate") {
      return true;
    }
    const taken = parent?.one_of?.[option];
    return (
      parent?.type === "identify" &&
      AUTHENTICATING_IDENTIFICATIONS.has(taken?.identification ?? "")
    );
  });
}

function fieldChecks(
  form: Form,
  fillable: ReadonlySet<string>,
  option?: number
): StepCheck[] {
  const out: StepCheck[] = [];
  const seen = new Set<string>();
  for (const field of form?.fields ?? []) {
    const pointer = field.user_profile?.pointer;
    if (pointer == null) {
      continue;
    }
    const at = option != null ? { option } : {};
    if (!fillable.has(pointer)) {
      out.push({ kind: "field_not_fillable", pointer, ...at });
    }
    if (seen.has(pointer)) {
      out.push({ kind: "field_duplicate", pointer, ...at });
    }
    seen.add(pointer);
  }
  return out;
}

function walk(
  steps: Step[] | undefined,
  pointer: string,
  visit: (step: Step, pointer: string) => void
) {
  (steps ?? []).forEach((step, i) => {
    const p = `${pointer}/${i}`;
    visit(step, p);
    (step.one_of ?? []).forEach((o, j) =>
      walk(o.steps, `${p}/one_of/${j}/steps`, visit)
    );
  });
}

// Problems with forms, by step. An option's form counts against its step.
export function stepChecks(
  flowType: FlowType,
  flow: AuthenticationFlowConfig,
  fillable: ReadonlySet<string>
): Map<string, StepCheck[]> {
  const out = new Map<string, StepCheck[]>();
  if (!formFlowTypes.has(flowType)) {
    return out;
  }
  walk(flow.steps, "/steps", (step, pointer) => {
    // What the server rejects comes first, then warnings.
    const checks: StepCheck[] = [];
    if (step.type === "fill_form") {
      const identified = identifiedBefore(flow, pointer);
      if (!identified) {
        checks.push({ kind: "form_before_identify" });
      }
      checks.push(...fieldChecks(step.form, fillable));
      if (
        identified &&
        flowType === "login" &&
        !authenticatedBefore(flow, pointer)
      ) {
        checks.push({ kind: "form_before_authentication" });
      }
    }
    (step.one_of ?? []).forEach((o, j) => {
      checks.push(...fieldChecks(o.form, fillable, j));
    });
    if (checks.length > 0) {
      out.set(pointer, checks);
    }
  });
  return out;
}

function hasStep(flow: AuthenticationFlowConfig, type: string): boolean {
  let found = false;
  walk(flow.steps, "/steps", (step) => {
    found ||= step.type === type;
  });
  return found;
}

// Problems with the flow as a whole.
export function flowChecks(
  flowType: FlowType,
  flow: AuthenticationFlowConfig
): FlowCheck[] {
  // Without the step, a disabled or expired account still completes the
  // login steps; it is stopped later, without being told why.
  if (flowType === "login" && !hasStep(flow, "check_account_status")) {
    return ["no_check_account_status"];
  }
  return [];
}

// Where generated login flows have check_account_status: right after the
// user has signed in.
export function checkAccountStatusAt(
  flow: AuthenticationFlowConfig
): InsertionPoint {
  const steps = flow.steps ?? [];
  let index = 0;
  steps.forEach((s, i) => {
    if (s.type === "identify" || s.type === "authenticate") {
      index = i + 1;
    }
  });
  return { steps: "/steps", index };
}
