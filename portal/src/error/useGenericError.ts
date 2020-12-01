import { useContext } from "react";
import { GraphQLError } from "graphql";
import { Context } from "@oursky/react-messageformat";

import { APIError, isAPIError, isApolloError } from "./error";
import { isLocationMatchWithJSONPointer } from "./useValidationError";
import { APIPasswordPolicyViolatedError } from "./password";
import {
  APIValidationError,
  ValidationFailedErrorInfoCause,
} from "./validation";
import { APIInvariantViolationError } from "./invariant";

interface BaseErrorHandlingRule {
  errorMessageID: string;
  field?: string;
}

interface ValidationErrorHandlingRule extends BaseErrorHandlingRule {
  reason: "ValidationFailed";
  kind: ValidationFailedErrorInfoCause["kind"];
  jsonPointer: string | RegExp;
}

interface InvariantViolationErrorHandlingRule extends BaseErrorHandlingRule {
  reason: "InvariantViolated";
  kind: string;
}

interface PasswordPolicyViolationErrorHandlingRule
  extends BaseErrorHandlingRule {
  reason: "PasswordPolicyViolated";
  cause: string;
}

interface OtherErrorHandlingRule extends BaseErrorHandlingRule {
  reason: Exclude<
    APIError["reason"],
    "ValidationFailed" | "InvariantViolated" | "PasswordPolicyViolated"
  >;
}

export type GenericErrorHandlingRule =
  | ValidationErrorHandlingRule
  | InvariantViolationErrorHandlingRule
  | PasswordPolicyViolationErrorHandlingRule
  | OtherErrorHandlingRule;

function matchInvariantViolationErrorWithRule(
  extensions: APIInvariantViolationError,
  rule: InvariantViolationErrorHandlingRule
): { isMatch: boolean } {
  const cause = extensions.info.cause;
  if (cause.kind === rule.kind) {
    return { isMatch: true };
  }
  return { isMatch: false };
}

function matchPasswordPolicyViolationErrorWithRule(
  extensions: APIPasswordPolicyViolatedError,
  rule: PasswordPolicyViolationErrorHandlingRule
): { isMatch: boolean } {
  const causes = extensions.info.causes;
  const matchedCause = causes.find((cause) => cause.Name === rule.cause);
  if (matchedCause != null) {
    return { isMatch: true };
  }
  return { isMatch: false };
}

function matchValidationErrorWithRule(
  extensions: APIValidationError,
  rule: ValidationErrorHandlingRule
): { isMatch: boolean; cause?: ValidationFailedErrorInfoCause } {
  const causes = extensions.info.causes;
  for (const cause of causes) {
    if (
      rule.kind === cause.kind &&
      isLocationMatchWithJSONPointer(rule.jsonPointer, cause.location)
    ) {
      return { isMatch: true, cause };
    }
  }
  return { isMatch: false };
}

function matchAPIErrorWithRule(
  extensions: APIError,
  rule: GenericErrorHandlingRule
): {
  isMatch: boolean;
  cause?: ValidationFailedErrorInfoCause;
} {
  // some error reason need special handling
  // depends on error info
  if (
    extensions.reason === "InvariantViolated" &&
    rule.reason === "InvariantViolated"
  ) {
    return matchInvariantViolationErrorWithRule(extensions, rule);
  }
  if (
    extensions.reason === "PasswordPolicyViolated" &&
    rule.reason === "PasswordPolicyViolated"
  ) {
    return matchPasswordPolicyViolationErrorWithRule(extensions, rule);
  }
  if (
    extensions.reason === "ValidationFailed" &&
    rule.reason === "ValidationFailed"
  ) {
    return matchValidationErrorWithRule(extensions, rule);
  }
  // for other error reason, only need to match reason
  return { isMatch: extensions.reason === rule.reason };
}

function constructErrorMessageFromGenericGraphQLError(
  error: GraphQLError,
  rules: GenericErrorHandlingRule[]
): {
  errorMessageId: string;
  violatedRule: GenericErrorHandlingRule;
  cause?: ValidationFailedErrorInfoCause;
} | null {
  if (!isAPIError(error.extensions)) {
    return null;
  }

  const { extensions } = error;
  for (const rule of rules) {
    const { isMatch, cause } = matchAPIErrorWithRule(extensions, rule);
    if (isMatch) {
      return {
        errorMessageId: rule.errorMessageID,
        violatedRule: rule,
        cause,
      };
    }
  }

  // no matching rules
  return null;
}

interface ErrorParseResult {
  standaloneErrorMessageIds: string[];
  fieldErrorMessageIds: Record<string, string[]>;
  unrecognizedError?: unknown;
  unhandledCauses: ValidationFailedErrorInfoCause[];
}

export function parseError(
  error: unknown,
  unhandledCauses: ValidationFailedErrorInfoCause[] | undefined,
  rules: GenericErrorHandlingRule[]
): ErrorParseResult {
  if (!isApolloError(error)) {
    console.warn("[Handle generic error]: Unhandled error\n", error);
    return {
      standaloneErrorMessageIds: [],
      fieldErrorMessageIds: {},
      unrecognizedError: error,
      unhandledCauses: [],
    };
  }

  const standaloneErrorMessageIds: string[] = [];
  const fieldErrorMessageIds: Partial<Record<string, string[]>> = {};
  const matchedCauses: ValidationFailedErrorInfoCause[] = [];
  let containUnrecognizedError = false;
  for (const graphQLError of error.graphQLErrors) {
    const violation = constructErrorMessageFromGenericGraphQLError(
      graphQLError,
      rules
    );
    if (violation != null) {
      const { errorMessageId, violatedRule, cause } = violation;
      if (violatedRule.field != null) {
        fieldErrorMessageIds[violatedRule.field] = [
          ...(fieldErrorMessageIds[violatedRule.field] ?? []),
          errorMessageId,
        ];
      } else {
        standaloneErrorMessageIds.push(errorMessageId);
      }
      if (cause != null) {
        matchedCauses.push(cause);
      }
    } else {
      console.warn(
        "[Handle generic error]: Contains unrecognized graphQL error \n",
        graphQLError
      );
      containUnrecognizedError = true;
    }
  }

  const filteredUnhandledCauses = unhandledCauses?.filter((cause) => {
    for (const matchedCause of matchedCauses) {
      if (
        matchedCause.kind === cause.kind &&
        matchedCause.location === cause.location
      ) {
        return false;
      }
    }
    return true;
  });

  return {
    standaloneErrorMessageIds,
    fieldErrorMessageIds: fieldErrorMessageIds as Record<string, string[]>,
    unrecognizedError: containUnrecognizedError ? error : undefined,
    unhandledCauses: filteredUnhandledCauses ?? [],
  };
}

export function useGenericError(
  error: unknown,
  unhandledCauses: ValidationFailedErrorInfoCause[] | undefined,
  rules: GenericErrorHandlingRule[],
  fallbackErrorMessageID: string = "generic-error.unknown-error"
): {
  errorMessage: string | undefined;
  errorMessageMap: Partial<Record<string, string>>;
  unrecognizedError?: unknown;
  unhandledCauses?: ValidationFailedErrorInfoCause[];
} {
  const { renderToString } = useContext(Context);

  if (error == null) {
    return { errorMessage: undefined, errorMessageMap: {} };
  }

  const result = parseError(error, unhandledCauses, rules);

  const errorMessageIds = result.standaloneErrorMessageIds.slice();
  const errorMessageMap: Record<string, string> = {};
  for (const [field, messageIds] of Object.entries(
    result.fieldErrorMessageIds
  )) {
    errorMessageMap[field] = messageIds
      .map((id) => renderToString(id))
      .join("\n");
    errorMessageIds.push(...messageIds);
  }
  if (result.unrecognizedError) {
    errorMessageIds.push(fallbackErrorMessageID);
  }

  return {
    errorMessage: errorMessageIds.map((id) => renderToString(id)).join("\n"),
    errorMessageMap,
    unrecognizedError: result.unrecognizedError,
    unhandledCauses: result.unhandledCauses,
  };
}
