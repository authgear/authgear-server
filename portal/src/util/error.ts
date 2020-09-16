import { ApolloError } from "@apollo/client";
import { GraphQLError } from "graphql";
import { nonNullable } from "./types";
import { Violation } from "./validation";

// expected data shape of error extension from backend
interface RequiredErrorCauseDetails {
  actual: string[];
  expected: string[];
  missing: string[];
}

interface RequiredErrorCause {
  details: RequiredErrorCauseDetails;
  location: string;
  kind: "required";
}

interface GeneralErrorCauseDetails {
  msg: string;
}

interface GeneralErrorCause {
  details: GeneralErrorCauseDetails;
  location: string;
  kind: "general";
}

// union type of cause details, depend on kind
type ErrorCause = RequiredErrorCause | GeneralErrorCause;

interface ValidationErrorInfo {
  causes: ErrorCause[];
}

interface APIValidationError {
  errorName: string;
  info: ValidationErrorInfo;
  reason: "ValidationFailed";
}

// union type of api errors, depend on reason
type APIError = APIValidationError;

function isAPIError(value?: { [key: string]: any }): value is APIError {
  if (value == null) {
    return false;
  }
  return "errorName" in value && "info" in value && "reason" in value;
}

function extractViolationFromErrorCause(cause: ErrorCause): Violation | null {
  switch (cause.kind) {
    case "required":
      return {
        kind: cause.kind,
        missingField: cause.details.missing,
        location: cause.location,
      };
    case "general":
      return {
        kind: cause.kind,
        location: cause.location,
      };
    default:
      return null;
  }
}

function handleUpdateAppConfigError(error: GraphQLError): Violation[] {
  if (!isAPIError(error.extensions)) {
    return [];
  }
  const causes = error.extensions.info.causes;
  /* uncomment when there is more than one error reason
  if (error.extensions.reason !== "ValidationFailed") {
    return [];
  }
  */

  return causes.map(extractViolationFromErrorCause).filter(nonNullable);
}

export function parseError(error: unknown): Violation[] {
  if (error instanceof ApolloError) {
    const violations: Violation[] = [];
    for (const graphQLError of error.graphQLErrors) {
      const errorViolations = handleUpdateAppConfigError(graphQLError);
      for (const violation of errorViolations) {
        violations.push(violation);
      }
    }
    return violations;
  }

  // unrecognized error
  return [];
}
