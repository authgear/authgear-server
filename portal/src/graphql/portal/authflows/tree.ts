import {
  AuthenticationFlowConfig,
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
} from "../../../types";

export type StepKind = "page" | "pageless" | "unknown";

export interface FlowBranch {
  optionIndex: number;
  option: AuthenticationFlowOptionConfig;
  sequence: FlowStepItem[];
}

export interface FlowStepItem {
  // JSON pointer of the step within its flow.
  id: string;
  step: AuthenticationFlowStepConfig;
  kind: StepKind;
  optional: boolean;
  // null: not a choice. []: a choice with no options.
  branches: FlowBranch[] | null;
}

export const IDENTIFICATION_ID_TOKEN = "id_token";

const pageStepTypes = new Set([
  "identify",
  "authenticate",
  "create_authenticator",
  "verify",
  "fill_in_user_profile",
  "fill_form",
  "view_recovery_code",
  "prompt_create_passkey",
  "change_password",
  "select_destination",
  "verify_account_recovery_code",
  "reset_password",
]);

const pagelessStepTypes = new Set([
  "check_account_status",
  "terminate_other_sessions",
]);

// Steps whose options are required; a missing one_of is a choice with none.
const choiceStepTypes = new Set([
  "identify",
  "authenticate",
  "create_authenticator",
]);

function stepKind(step: AuthenticationFlowStepConfig): StepKind {
  if (pagelessStepTypes.has(step.type)) {
    return "pageless";
  }
  const oneOf = step.one_of ?? [];
  if (
    step.type === "identify" &&
    oneOf.length > 0 &&
    oneOf.every((o) => o.identification === IDENTIFICATION_ID_TOKEN)
  ) {
    return "pageless";
  }
  if (pageStepTypes.has(step.type)) {
    return "page";
  }
  return "unknown";
}

function sequenceToTree(
  steps: AuthenticationFlowStepConfig[],
  pointer: string
): FlowStepItem[] {
  const items: FlowStepItem[] = [];
  steps.forEach((step, i) => {
    const id = `${pointer}/${i}`;
    const oneOf = step.one_of ?? [];
    const optional = step.optional === true;
    const kind = stepKind(step);
    // An optional step keeps its single option as a branch, so skipping it
    // also skips the option's nested steps.
    if (oneOf.length === 1 && !optional) {
      items.push({ id, step, kind, optional, branches: null });
      items.push(
        ...sequenceToTree(oneOf[0].steps ?? [], `${id}/one_of/0/steps`)
      );
      return;
    }
    if (oneOf.length === 0 && !choiceStepTypes.has(step.type)) {
      items.push({ id, step, kind, optional, branches: null });
      return;
    }
    items.push({
      id,
      step,
      kind,
      optional,
      branches: oneOf.map((option, j) => ({
        optionIndex: j,
        option,
        sequence: sequenceToTree(option.steps ?? [], `${id}/one_of/${j}/steps`),
      })),
    });
  });
  return items;
}

export function flowToTree(flow: AuthenticationFlowConfig): FlowStepItem[] {
  return sequenceToTree(flow.steps ?? [], "/steps");
}
