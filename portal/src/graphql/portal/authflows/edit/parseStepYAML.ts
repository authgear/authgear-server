import yaml from "js-yaml";
import { AuthenticationFlowStepConfig } from "../../../../types";
import { FlowType } from "../model";
import { allowedStepTypes } from "./rules";

export type ParseStepYAMLError = "parse" | "not_step" | "type";

export type ParseStepYAMLResult =
  | { step: AuthenticationFlowStepConfig }
  | { error: ParseStepYAMLError };

// fill_in_user_profile is not offered for new steps, but an existing one
// stays editable here.
const DEPRECATED_STEP_TYPES: Partial<Record<FlowType, string[]>> = {
  signup: ["fill_in_user_profile"],
  promote: ["fill_in_user_profile"],
};

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
  return { step };
}
