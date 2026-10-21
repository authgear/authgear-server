import { useContext, useMemo } from "react";
import { Context } from "../../../../intl";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowStepConfig,
} from "../../../../types";
import { FlowType } from "../model";
import { optionLabel, stepTitle } from "../labels";
import {
  FlowEditorContextValue,
  InsertableType,
  StepProblem,
} from "./FlowEditorContext";
import { EditResult, insertStep } from "./edits";
import { MoveError, planMove } from "./move";
import {
  InsertionPoint,
  getAt,
  optionOf,
  parentSteps,
  stepPointer,
} from "./pointer";
import { isDeviceTokenOption } from "../tree";
import {
  ReferenceWarning,
  TargetUsage,
  ensureNamed,
  referenceWarnings,
  removalBlockers,
  targetCandidates,
} from "./references";
import { TARGET_STEP_TYPES, allowedStepTypes, newStep } from "./rules";
import { identifiedBefore } from "./checks";

const moveErrorMessageIDs: Record<MoveError, string> = {
  into_itself: "AuthFlowsScreen.edit.move.error.into-itself",
  option_needs_a_step: "AuthFlowsScreen.edit.blocked.only-nested-step",
  not_allowed_here: "AuthFlowsScreen.edit.move.error.not-allowed-here",
  other_option_has_copy:
    "AuthFlowsScreen.edit.move.error.other-option-has-copy",
};

const warningMessageIDs: Record<ReferenceWarning, string> = {
  missing: "AuthFlowsScreen.edit.warning.missing",
  wrong_type: "AuthFlowsScreen.edit.warning.wrong_type",
  not_earlier: "AuthFlowsScreen.edit.warning.not_earlier",
  wrong_contact: "AuthFlowsScreen.edit.warning.wrong_contact",
};

const noTargetMessageIDs: Partial<Record<string, string>> = {
  verify: "AuthFlowsScreen.edit.insert.no-identify",
  change_password: "AuthFlowsScreen.edit.insert.no-authenticate",
};

function isTargetUsage(type: string): type is TargetUsage {
  return type === "verify" || type === "change_password";
}

export interface FlowEditorOptions {
  flowType: FlowType;
  flow: AuthenticationFlowConfig;
  // Nodes with errors from the last save.
  errorStepIDs: ReadonlySet<string>;
  // Each node's first save error or check; target_step warnings come after.
  problems: ReadonlyMap<string, StepProblem>;
  // Attributes a form can ask for.
  fillablePointers: readonly string[];
  // Applies an edit; notice is shown in the panel of the step it selects.
  apply: (result: EditResult, notice?: string) => void;
  // Removes a step, after confirming when steps are nested under it.
  requestRemove: (stepID: string) => void;
  // Shows why a dropped card did not move.
  notify: (message: string) => void;
}

export function useFlowEditorValue({
  flowType,
  flow,
  errorStepIDs,
  problems,
  fillablePointers,
  apply,
  requestRemove,
  notify,
}: FlowEditorOptions): FlowEditorContextValue {
  const { locale, renderToString } = useContext(Context);
  const warnings = useMemo(
    () => referenceWarnings(flowType, flow),
    [flowType, flow]
  );

  return useMemo<FlowEditorContextValue>(() => {
    const stepLabel = (pointer: string) => {
      const step = getAt(flow, pointer) as { type: string; name?: string };
      return step.name ?? stepTitle(step.type, renderToString);
    };

    const insertableTypes = (at: InsertionPoint): InsertableType[] => {
      // signup_login only has identify steps, at the top level.
      if (flowType === "signup_login" && at.steps !== "/steps") {
        return [];
      }
      return allowedStepTypes(flowType).map((type) => {
        if (type === "fill_form") {
          // The server rejects a form before any identify step, and a form
          // needs a field.
          const disabledReason = !identifiedBefore(flow, stepPointer(at))
            ? "AuthFlowsScreen.edit.insert.no-identify-before-form"
            : fillablePointers.length === 0
            ? "AuthFlowsScreen.edit.insert.no-fillable"
            : null;
          return { type, disabledReason };
        }
        const reason = noTargetMessageIDs[type];
        const blocked =
          isTargetUsage(type) &&
          targetCandidates(flow, stepPointer(at), type).length === 0;
        return { type, disabledReason: blocked ? reason ?? null : null };
      });
    };

    const onInsert = (at: InsertionPoint, type: string) => {
      let next = flow;
      let target: string | null = null;
      let notice: string | undefined;
      if (isTargetUsage(type)) {
        const candidates = targetCandidates(flow, stepPointer(at), type);
        const nearest = candidates[candidates.length - 1];
        const targetType = TARGET_STEP_TYPES[type] ?? type;
        const named = ensureNamed(flow, nearest.pointer, targetType);
        if (nearest.name == null) {
          notice = renderToString("AuthFlowsScreen.edit.notice.named", {
            name: named.name,
          });
        }
        next = named.flow;
        target = named.name;
      }
      apply(
        insertStep(
          next,
          at,
          newStep(flowType, type, target, fillablePointers[0] ?? null)
        ),
        notice
      );
    };

    // Called while rendering nodes, which may briefly name steps from before
    // an edit, so a missing step is not an error.
    const removeBlockedReason = (stepID: string): string | null => {
      if (stepID === "/steps/0" && (flow.steps ?? []).length === 1) {
        return renderToString("AuthFlowsScreen.edit.blocked.only-step");
      }
      // An account recovery option needs at least one nested step.
      const parent = parentSteps(stepID);
      if (
        flowType === "account_recovery" &&
        parent.steps !== "/steps" &&
        (getAt(flow, parent.steps) as unknown[] | undefined)?.length === 1
      ) {
        return renderToString("AuthFlowsScreen.edit.blocked.only-nested-step");
      }
      const blockers = removalBlockers(flow, stepID);
      if (blockers.length > 0) {
        return renderToString("AuthFlowsScreen.edit.blocked.targeted", {
          steps: blockers.map(stepLabel).join(", "),
        });
      }
      return null;
    };

    const removeOptionBlockedReason = (
      stepID: string,
      index: number
    ): string | null => {
      const options =
        (getAt(flow, stepID) as AuthenticationFlowStepConfig).one_of ?? [];
      const choices = options.filter((o) => !isDeviceTokenOption(o));
      if (!isDeviceTokenOption(options[index]) && choices.length <= 1) {
        return renderToString("AuthFlowsScreen.edit.blocked.only-option");
      }
      const blockers = removalBlockers(flow, `${stepID}/one_of/${index}`);
      if (blockers.length > 0) {
        return renderToString("AuthFlowsScreen.edit.blocked.targeted", {
          steps: blockers.map(stepLabel).join(", "),
        });
      }
      return null;
    };

    const onMove = (stepID: string, to: InsertionPoint, copy: boolean) => {
      const p = planMove(flowType, flow, stepID, to, copy);
      if (!("error" in p)) {
        if (p.flow !== flow) {
          apply(p);
        }
        return;
      }
      if (p.error !== "other_option_has_copy") {
        notify(renderToString(moveErrorMessageIDs[p.error]));
        return;
      }
      const options =
        (getAt(flow, p.choice) as AuthenticationFlowStepConfig).one_of ?? [];
      const step = getAt(flow, stepID) as AuthenticationFlowStepConfig;
      notify(
        renderToString(moveErrorMessageIDs[p.error], {
          options: new Intl.ListFormat(locale).format(
            p.options.map((j) => optionLabel(options[j], renderToString))
          ),
          count: p.options.length,
          step: stepTitle(step.type, renderToString),
        })
      );
    };

    const optionAt = (at: InsertionPoint) => {
      const o = optionOf(at);
      if (o == null) {
        return null;
      }
      const step = getAt(flow, o.choice) as
        | AuthenticationFlowStepConfig
        | undefined;
      return step?.one_of?.[o.optionIndex] ?? null;
    };

    return {
      insertableTypes,
      optionAt,
      onMove,
      removeOptionBlockedReason,
      onInsert,
      removeBlockedReason,
      onRemove: requestRemove,
      hasErrors: (stepID) => errorStepIDs.has(stepID),
      problem: (stepID) => {
        const p = problems.get(stepID);
        if (p != null) {
          return p;
        }
        const w = warnings.get(stepID);
        return w != null
          ? { tone: "warning", text: renderToString(warningMessageIDs[w]) }
          : null;
      },
    };
  }, [
    flowType,
    flow,
    errorStepIDs,
    problems,
    fillablePointers,
    apply,
    requestRemove,
    notify,
    warnings,
    renderToString,
    locale,
  ]);
}
