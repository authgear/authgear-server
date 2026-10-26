import { produce } from "immer";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowStepConfig,
} from "../../../../types";
import { FlowType } from "../model";
import { isDeviceTokenOption } from "../tree";
import { EditResult } from "./edits";
import {
  InsertionPoint,
  getAt,
  isWithin,
  parentSteps,
  stepPointer,
} from "./pointer";

// Moving or copying a step to another place, as dragging a card onto a "+"
// does. Where a step sits says which options run it: inside an option only
// that option, after a choice every option.

type Step = AuthenticationFlowStepConfig;

export type MoveError =
  | "into_itself"
  | "option_needs_a_step"
  | "not_allowed_here"
  | "other_option_has_copy";

export type MovePlan =
  | { error: Exclude<MoveError, "other_option_has_copy"> }
  // Moving the step after choice would run it twice on these options.
  | { error: "other_option_has_copy"; choice: string; options: number[] }
  | EditResult;

// What a choice step's options set up: its identifications or authenticators.
function optionKinds(step: Step): Set<string> {
  return new Set(
    (step.one_of ?? []).flatMap((o) =>
      isDeviceTokenOption(o) ? [] : [o.identification ?? o.authentication ?? ""]
    )
  );
}

// Whether other does what step does: the same type and, for a choice, at
// least one identification or authenticator in common. A password step and
// a 2FA step are both create_authenticator, but neither repeats the other.
function isCopy(other: Step, step: Step): boolean {
  if (other.type !== step.type) {
    return false;
  }
  const kinds = optionKinds(step);
  if (kinds.size === 0) {
    return true;
  }
  return [...optionKinds(other)].some((k) => kinds.has(k));
}

function containsCopy(steps: Step[] | undefined, step: Step): boolean {
  return (steps ?? []).some(
    (s) =>
      isCopy(s, step) ||
      (s.one_of ?? []).some((o) => containsCopy(o.steps, step))
  );
}

// The choices whose options lead to the step at pointer, innermost first,
// with the option it is in.
function enclosingOptions(
  pointer: string
): { choice: string; optionIndex: number }[] {
  const out: { choice: string; optionIndex: number }[] = [];
  const re = /\/one_of\/(\d+)\/steps\/\d+$/;
  let p = pointer;
  for (;;) {
    const m = re.exec(p);
    if (m == null) {
      return out;
    }
    const choice = p.slice(0, m.index);
    out.push({ choice, optionIndex: Number(m[1]) });
    p = choice;
  }
}

// A pointer after the step at steps/index is taken out of its list.
function afterRemoval(pointer: string, steps: string, index: number): string {
  const prefix = `${steps}/`;
  if (!pointer.startsWith(prefix)) {
    return pointer;
  }
  const rest = pointer.slice(prefix.length);
  const slash = rest.indexOf("/");
  const i = Number(slash < 0 ? rest : rest.slice(0, slash));
  const tail = slash < 0 ? "" : rest.slice(slash);
  return i > index ? `${prefix}${i - 1}${tail}` : pointer;
}

function withoutNames(step: Step): Step {
  return produce(step, (d) => {
    const strip = (s: Step) => {
      delete s.name;
      for (const o of s.one_of ?? []) {
        (o.steps ?? []).forEach(strip);
      }
    };
    strip(d);
  });
}

export function planMove(
  flowType: FlowType,
  flow: AuthenticationFlowConfig,
  from: string,
  to: InsertionPoint,
  copy: boolean
): MovePlan {
  if (isWithin(to.steps, from)) {
    return { error: "into_itself" };
  }
  // signup_login only has identify steps, at the top level.
  if (flowType === "signup_login" && to.steps !== "/steps") {
    return { error: "not_allowed_here" };
  }
  const source = parentSteps(from);
  if (
    !copy &&
    to.steps === source.steps &&
    (to.index === source.index || to.index === source.index + 1)
  ) {
    return { flow, selected: from };
  }
  const sourceList = getAt(flow, source.steps) as Step[];
  const step = sourceList[source.index];
  if (
    !copy &&
    flowType === "account_recovery" &&
    source.steps !== "/steps" &&
    to.steps !== source.steps &&
    sourceList.length === 1
  ) {
    return { error: "option_needs_a_step" };
  }
  // Out of a choice to after it: the other options then run it too.
  for (const { choice, optionIndex } of enclosingOptions(from)) {
    const at = parentSteps(choice);
    if (to.steps !== at.steps || to.index <= at.index) {
      continue;
    }
    const options = (getAt(flow, choice) as Step).one_of ?? [];
    const copies = options.flatMap((o, j) =>
      j !== optionIndex &&
      !isDeviceTokenOption(o) &&
      containsCopy(o.steps, step)
        ? [j]
        : []
    );
    if (copies.length > 0) {
      return { error: "other_option_has_copy", choice, options: copies };
    }
  }

  let target = to;
  if (!copy) {
    target = {
      steps: afterRemoval(to.steps, source.steps, source.index),
      index:
        to.steps === source.steps && to.index > source.index
          ? to.index - 1
          : to.index,
    };
  }
  const next = produce(flow, (d) => {
    const list = getAt(d, source.steps) as Step[];
    // The owner of the source list stays the same object while indexes shift.
    const slash = source.steps.lastIndexOf("/");
    const owner = getAt(d, source.steps.slice(0, slash)) as Record<
      string,
      unknown
    >;
    if (!copy) {
      list.splice(source.index, 1);
    }
    const tSlash = target.steps.lastIndexOf("/");
    const tOwner = getAt(d, target.steps.slice(0, tSlash)) as Record<
      string,
      unknown
    >;
    const key = target.steps.slice(tSlash + 1);
    tOwner[key] ??= [];
    (tOwner[key] as Step[]).splice(
      target.index,
      0,
      copy ? withoutNames(step) : step
    );
    if (list.length === 0 && source.steps !== "/steps") {
      delete owner[source.steps.slice(slash + 1)];
    }
  });
  return { flow: next, selected: stepPointer(target) };
}
