import { ValidationFailedErrorInfoCause } from "../../../../error/validation";
import { APIError } from "../../../../error/error";
import { ErrorParseRule, ErrorParseRuleResult } from "../../../../error/parse";
import { isWithin } from "./pointer";

// A save error on a step, as a message: a config path relative to the step
// (form.fields) at most, never the JSON pointer into authgear.yaml.

export interface CauseMessage {
  messageID: string;
  values?: Record<string, string | number>;
}

// The server's own messages for flows, by its exact text.
const generalMessageIDs: Partial<Record<string, string>> = {
  "fill_form must come after an identify step":
    "AuthFlowsScreen.check.form-before-identify",
  "attribute is not fillable": "AuthFlowsScreen.error.not-fillable",
  "duplicated pointer": "AuthFlowsScreen.error.duplicate-field",
  "target_step does not refer to any named step in the same flow":
    "AuthFlowsScreen.error.target-missing",
};

// location below stepLocation, as a dotted config path; "" for the step.
function settingPath(location: string, stepLocation: string): string {
  return location
    .slice(stepLocation.length)
    .split("/")
    .filter((t) => t !== "")
    .join(".");
}

export function describeCause(
  cause: ValidationFailedErrorInfoCause,
  stepLocation: string
): CauseMessage {
  if (cause.kind === "general") {
    const messageID = generalMessageIDs[cause.details.msg];
    return messageID != null
      ? { messageID }
      : {
          messageID: "AuthFlowsScreen.error.server",
          values: { message: cause.details.msg },
        };
  }
  const field = settingPath(cause.location, stepLocation);
  if (cause.kind === "required") {
    return {
      messageID: "AuthFlowsScreen.error.required",
      values: {
        fields: cause.details.missing
          .map((m) => (field === "" ? m : `${field}.${m}`))
          .join(", "),
      },
    };
  }
  if (cause.kind === "minItems") {
    return {
      messageID: "AuthFlowsScreen.error.min-items",
      values: { field, count: cause.details.expected },
    };
  }
  return field === ""
    ? { messageID: "AuthFlowsScreen.error.invalid-step" }
    : { messageID: "AuthFlowsScreen.error.invalid", values: { field } };
}

// For the error bar: the causes inside the flow are shown on its steps, so
// the bar only counts them. Other causes go through as usual.
export function flowErrorRule(prefix: string): ErrorParseRule {
  return (apiError: APIError): ErrorParseRuleResult => {
    if (apiError.reason !== "ValidationFailed") {
      return { parsedAPIErrors: [], fullyHandled: false };
    }
    const inFlow = apiError.info.causes.filter((c) =>
      isWithin(c.location, prefix)
    );
    if (inFlow.length === 0) {
      return { parsedAPIErrors: [], fullyHandled: false };
    }
    const rest = apiError.info.causes.filter(
      (c) => !isWithin(c.location, prefix)
    );
    return {
      parsedAPIErrors: [
        {
          messageID: "AuthFlowsScreen.error.on-flow",
          arguments: { count: inFlow.length },
        },
      ],
      modifiedAPIError: {
        ...apiError,
        info: { ...apiError.info, causes: rest },
      },
      fullyHandled: rest.length === 0,
    };
  };
}
