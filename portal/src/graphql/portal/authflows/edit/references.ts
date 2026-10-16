import { produce } from "immer";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowStepConfig,
} from "../../../../types";
import { IDENTIFICATION_ID_TOKEN } from "../tree";
import { FlowType } from "../model";
import { EditResult } from "./edits";
import { getAt, isWithin, parsePointer } from "./pointer";

// target_step names another step of the same flow. The server checks on save
// that the name exists; the runtime checks the target's type and that it ran
// earlier (pkg/lib/authenticationflow/target_step.go). With duplicate names
// the later step wins.

type Step = AuthenticationFlowStepConfig;

export type TargetUsage =
  | "verify"
  | "create_authenticator"
  | "authenticate"
  | "change_password";

const usageTargetTypes: Record<TargetUsage, string> = {
  verify: "identify",
  create_authenticator: "identify",
  authenticate: "identify",
  change_password: "authenticate",
};

// Flow types whose steps can hold target_step.
const targetingFlowTypes = new Set<FlowType>(["signup", "promote", "login"]);

interface StepAt {
  pointer: string;
  step: Step;
}

function walk(steps: Step[] | undefined, pointer: string, out: StepAt[]) {
  (steps ?? []).forEach((step, i) => {
    const p = `${pointer}/${i}`;
    out.push({ pointer: p, step });
    (step.one_of ?? []).forEach((o, j) => {
      walk(o.steps, `${p}/one_of/${j}/steps`, out);
    });
  });
}

// Every step in document order, nested ones after their parent.
function allSteps(flow: AuthenticationFlowConfig): StepAt[] {
  const out: StepAt[] = [];
  walk(flow.steps, "/steps", out);
  return out;
}

function stepAt(flow: AuthenticationFlowConfig, pointer: string): Step {
  return getAt(flow, pointer) as Step;
}

// Steps that have run before the step at pointer, on any path to it: its
// ancestors, and every earlier sibling at each level with what is nested
// under it.
export function earlierSteps(
  flow: AuthenticationFlowConfig,
  pointer: string
): string[] {
  const tokens = parsePointer(pointer);
  const out: string[] = [];
  let stepsPointer = "";
  // tokens: steps, i, (one_of, j, steps, i)*
  for (let t = 0; t < tokens.length; t += 2) {
    stepsPointer = `${stepsPointer}/${tokens[t]}`;
    if (tokens[t] === "one_of") {
      stepsPointer = `${stepsPointer}/${tokens[t + 1]}`;
      continue;
    }
    const index = Number(tokens[t + 1]);
    const siblings = getAt(flow, stepsPointer) as Step[] | undefined;
    const before: StepAt[] = [];
    walk(siblings?.slice(0, index), stepsPointer, before);
    out.push(...before.map((s) => s.pointer));
    const isLast = t + 2 >= tokens.length;
    if (!isLast) {
      out.push(`${stepsPointer}/${index}`);
    }
    stepsPointer = `${stepsPointer}/${index}`;
  }
  // A pointer from a stale render may name steps that are gone.
  return out.filter((p) => getAt(flow, p) != null);
}

function isAllIDToken(step: Step): boolean {
  return (step.one_of ?? []).every(
    (o) => o.identification === IDENTIFICATION_ID_TOKEN
  );
}

function canTarget(usage: TargetUsage, target: Step): boolean {
  if (target.type !== usageTargetTypes[usage]) {
    return false;
  }
  // An authenticate option needs an identity, which id_token does not give.
  return usage !== "authenticate" || !isAllIDToken(target);
}

export interface TargetCandidate {
  pointer: string;
  name: string | null;
}

export function targetCandidates(
  flow: AuthenticationFlowConfig,
  pointer: string,
  usage: TargetUsage
): TargetCandidate[] {
  return earlierSteps(flow, pointer)
    .map((p) => ({ pointer: p, step: stepAt(flow, p) }))
    .filter(({ step }) => canTarget(usage, step))
    .map(({ pointer: p, step }) => ({ pointer: p, name: step.name ?? null }));
}

interface Reference {
  // The step holding target_step, itself or on one of its options.
  holder: string;
  name: string;
  usage: TargetUsage;
}

function isTargetUsage(type: string): type is TargetUsage {
  return type in usageTargetTypes;
}

function allReferences(flow: AuthenticationFlowConfig): Reference[] {
  const refs: Reference[] = [];
  for (const { pointer, step } of allSteps(flow)) {
    if (!isTargetUsage(step.type)) {
      continue;
    }
    if (step.target_step != null) {
      refs.push({ holder: pointer, name: step.target_step, usage: step.type });
    }
    for (const o of step.one_of ?? []) {
      if (o.target_step != null) {
        refs.push({ holder: pointer, name: o.target_step, usage: step.type });
      }
    }
  }
  return refs;
}

function stepsNamed(flow: AuthenticationFlowConfig, name: string): string[] {
  return allSteps(flow)
    .filter((s) => s.step.name === name)
    .map((s) => s.pointer);
}

function unique(xs: string[]): string[] {
  return [...new Set(xs)];
}

// The steps whose target_step names the step at pointer.
export function referencesTo(
  flow: AuthenticationFlowConfig,
  pointer: string
): string[] {
  const name = stepAt(flow, pointer).name;
  if (name == null) {
    return [];
  }
  return unique(
    allReferences(flow)
      .filter((r) => r.name === name)
      .map((r) => r.holder)
  );
}

// The steps outside pointer's subtree that would lose their target if it
// were removed.
export function removalBlockers(
  flow: AuthenticationFlowConfig,
  pointer: string
): string[] {
  const steps = allSteps(flow);
  const inside = steps.filter((s) => isWithin(s.pointer, pointer));
  const outsideNames = new Set(
    steps.filter((s) => !isWithin(s.pointer, pointer)).map((s) => s.step.name)
  );
  const lost = new Set(
    inside
      .map((s) => s.step.name)
      .filter((n): n is string => n != null && !outsideNames.has(n))
  );
  return unique(
    allReferences(flow)
      .filter((r) => lost.has(r.name) && !isWithin(r.holder, pointer))
      .map((r) => r.holder)
  );
}

// Renames a step; "" removes the name. References follow when the old name
// was the step's alone; with a shared name they may resolve to the other step.
export function renameStep(
  flow: AuthenticationFlowConfig,
  pointer: string,
  name: string
): EditResult & { rewritten: boolean } {
  const old = stepAt(flow, pointer).name;
  const follow =
    old != null && name !== "" && stepsNamed(flow, old).length === 1;
  let rewritten = false;
  const next = produce(flow, (d) => {
    const step = getAt(d, pointer) as Step;
    if (name === "") {
      delete step.name;
    } else {
      step.name = name;
    }
    if (!follow) {
      return;
    }
    for (const { step: s } of allSteps(d)) {
      if (!isTargetUsage(s.type)) {
        continue;
      }
      if (s.target_step === old) {
        s.target_step = name;
        rewritten = true;
      }
      for (const o of s.one_of ?? []) {
        if (o.target_step === old) {
          o.target_step = name;
          rewritten = true;
        }
      }
    }
  });
  return { flow: next, selected: pointer, rewritten };
}

// Gives the step at pointer a name, if it has none, that no other step has.
export function ensureNamed(
  flow: AuthenticationFlowConfig,
  pointer: string,
  base: string
): { flow: AuthenticationFlowConfig; name: string } {
  const existing = stepAt(flow, pointer).name;
  if (existing != null) {
    return { flow, name: existing };
  }
  const taken = new Set(allSteps(flow).map((s) => s.step.name));
  let name = base;
  for (let n = 2; taken.has(name); n++) {
    name = `${base}_${n}`;
  }
  const next = produce(flow, (d) => {
    (getAt(d, pointer) as Step).name = name;
  });
  return { flow: next, name };
}

export type ReferenceWarning = "missing" | "wrong_type" | "not_earlier";

// Problems with each step's target_step, keyed by the step holding it.
// Only "missing" fails the save; the others fail at runtime.
export function referenceWarnings(
  flowType: FlowType,
  flow: AuthenticationFlowConfig
): Map<string, ReferenceWarning> {
  const warnings = new Map<string, ReferenceWarning>();
  if (!targetingFlowTypes.has(flowType)) {
    return warnings;
  }
  for (const ref of allReferences(flow)) {
    if (warnings.has(ref.holder)) {
      continue;
    }
    const named = stepsNamed(flow, ref.name);
    if (named.length === 0) {
      warnings.set(ref.holder, "missing");
      continue;
    }
    const target = named[named.length - 1];
    if (!canTarget(ref.usage, stepAt(flow, target))) {
      warnings.set(ref.holder, "wrong_type");
    } else if (!earlierSteps(flow, ref.holder).includes(target)) {
      warnings.set(ref.holder, "not_earlier");
    }
  }
  return warnings;
}

export function duplicateNames(flow: AuthenticationFlowConfig): Set<string> {
  const seen = new Set<string>();
  const dup = new Set<string>();
  for (const { step } of allSteps(flow)) {
    if (step.name == null) {
      continue;
    }
    if (seen.has(step.name)) {
      dup.add(step.name);
    }
    seen.add(step.name);
  }
  return dup;
}
