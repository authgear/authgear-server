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

type InvariantViolationErrorKind = "RemoveLastIdentity";

interface InvariantViolationErrorCause {
  kind: InvariantViolationErrorKind;
}

interface InvariantViolationErrorInfo {
  cause: InvariantViolationErrorCause;
}

interface APIInvariantViolationError {
  errorName: string;
  info: InvariantViolationErrorInfo;
  reason: "InvariantViolated";
}

// union type of api errors, depend on reason
type APIError = APIValidationError | APIInvariantViolationError;

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

export function handleUpdateAppConfigError(error: GraphQLError): Violation[] {
  if (!isAPIError(error.extensions)) {
    return [];
  }
  const { extensions } = error;
  switch (extensions.reason) {
    case "ValidationFailed": {
      const causes = extensions.info.causes;
      return causes.map(extractViolationFromErrorCause).filter(nonNullable);
    }
    case "InvariantViolated": {
      const cause = extensions.info.cause;
      return [{ kind: cause.kind }];
    }
    default:
      return [];
  }
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
