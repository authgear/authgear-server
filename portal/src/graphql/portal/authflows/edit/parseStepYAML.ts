import yaml from "js-yaml";
import { AuthenticationFlowStepConfig } from "../../../../types";
import { FlowType } from "../model";
import { allowedStepTypes } from "./rules";

export type ParseStepYAMLError = "parse" | "not_step" | "type" | "form";

export type ParseStepYAMLResult =
  | { step: AuthenticationFlowStepConfig }
  | { error: ParseStepYAMLError };

// fill_in_user_profile is not offered for new steps, but an existing one
// stays editable here.
const DEPRECATED_STEP_TYPES: Partial<Record<FlowType, string[]>> = {
  signup: ["fill_in_user_profile"],
  promote: ["fill_in_user_profile"],
};

// Where the server allows a form: on fill_form steps, and on the email,
// phone and username options of a signup identify step.
const formOptionIdentifications = new Set(["email", "phone", "username"]);

function misplacedForm(
  step: AuthenticationFlowStepConfig,
  flowType: FlowType
): boolean {
  if (step.form != null && step.type !== "fill_form") {
    return true;
  }
  return (step.one_of ?? []).some(
    (o) =>
      (o.form != null &&
        !(
          (flowType === "signup" || flowType === "promote") &&
          step.type === "identify" &&
          formOptionIdentifications.has(o.identification ?? "")
        )) ||
      (o.steps ?? []).some((s) => misplacedForm(s, flowType))
  );
}

export function parseStepYAML(
  text: string,
  flowType: FlowType
): ParseStepYAMLResult {
  let value: unknown;
  try {
    value = yaml.load(text);
  } catch {
    // The editor shows the parse error; the draft is left as it was.
    return { error: "parse" };
  }
  if (
    value == null ||
    typeof value !== "object" ||
    Array.isArray(value) ||
    typeof (value as { type?: unknown }).type !== "string"
  ) {
    return { error: "not_step" };
  }
  const step = value as AuthenticationFlowStepConfig;
  const allowed = [
    ...allowedStepTypes(flowType),
    ...(DEPRECATED_STEP_TYPES[flowType] ?? []),
  ];
  if (!allowed.includes(step.type)) {
    return { error: "type" };
  }
  if (misplacedForm(step, flowType)) {
    return { error: "form" };
  }
  return { step };
}
