import { produce, Draft } from "immer";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
} from "../../../../types";
import {
  InsertionPoint,
  getAt,
  parentSteps,
  parsePointer,
  stepPointer,
} from "./pointer";

// Every edit returns a new flow, sharing what it did not change, and the
// pointer of the step to select afterwards.
export interface EditResult {
  flow: AuthenticationFlowConfig;
  selected: string | null;
}

type Step = AuthenticationFlowStepConfig;

function draftAt(
  root: Draft<AuthenticationFlowConfig>,
  pointer: string
): unknown {
  const value = getAt(root, pointer);
  if (value == null) {
    throw new Error(`no value at ${pointer}`);
  }
  return value;
}

// The steps array at pointer, created when an option has none yet.
function draftSteps(
  root: Draft<AuthenticationFlowConfig>,
  pointer: string
): Step[] {
  const existing = getAt(root, pointer);
  if (existing != null) {
    return existing as Step[];
  }
  const tokens = parsePointer(pointer);
  const key = tokens[tokens.length - 1];
  const owner = draftAt(root, `/${tokens.slice(0, -1).join("/")}`) as Record<
    string,
    unknown
  >;
  const steps: Step[] = [];
  owner[key] = steps;
  return owner[key] as Step[];
}

export function insertStep(
  flow: AuthenticationFlowConfig,
  at: InsertionPoint,
  step: Step
): EditResult {
  return {
    flow: produce(flow, (d) => {
      draftSteps(d, at.steps).splice(at.index, 0, step);
    }),
    selected: stepPointer(at),
  };
}

export function removeStep(
  flow: AuthenticationFlowConfig,
  pointer: string
): EditResult {
  const { steps, index } = parentSteps(pointer);
  return {
    flow: produce(flow, (d) => {
      const list = draftAt(d, steps) as Step[];
      list.splice(index, 1);
      // An option left with no steps loses the key, as written by hand.
      if (list.length === 0 && steps !== "/steps") {
        const slash = steps.lastIndexOf("/");
        const owner = draftAt(d, steps.slice(0, slash)) as Record<
          string,
          unknown
        >;
        delete owner[steps.slice(slash + 1)];
      }
    }),
    selected: null,
  };
}

export function addOption(
  flow: AuthenticationFlowConfig,
  pointer: string,
  option: AuthenticationFlowOptionConfig
): EditResult {
  return {
    flow: produce(flow, (d) => {
      const step = draftAt(d, pointer) as Step;
      step.one_of = [...(step.one_of ?? []), option];
    }),
    selected: pointer,
  };
}

export function removeOption(
  flow: AuthenticationFlowConfig,
  pointer: string,
  index: number
): EditResult {
  return {
    flow: produce(flow, (d) => {
      (draftAt(d, pointer) as Step).one_of?.splice(index, 1);
    }),
    selected: pointer,
  };
}

export function updateStep(
  flow: AuthenticationFlowConfig,
  pointer: string,
  recipe: (step: Step) => void
): EditResult {
  return {
    flow: produce(flow, (d) => {
      recipe(draftAt(d, pointer) as Step);
    }),
    selected: pointer,
  };
}

export function replaceStep(
  flow: AuthenticationFlowConfig,
  pointer: string,
  step: Step
): EditResult {
  const { steps, index } = parentSteps(pointer);
  return {
    flow: produce(flow, (d) => {
      (draftAt(d, steps) as Step[])[index] = step;
    }),
    selected: pointer,
  };
}

// Steps nested under an option, at any depth.
export function countOptionSteps(
  option: AuthenticationFlowOptionConfig
): number {
  return (option.steps ?? []).reduce((n, s) => n + 1 + countNestedSteps(s), 0);
}

export function countNestedSteps(step: Step): number {
  return (step.one_of ?? []).reduce((n, o) => n + countOptionSteps(o), 0);
}
