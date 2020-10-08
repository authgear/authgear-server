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

interface FormatErrorCauseDetails {
  format: string;
}

interface FormatErrorCause {
  details: FormatErrorCauseDetails;
  location: string;
  kind: "format";
}

interface MinItemsErrorCauseDetails {
  actual: number;
  expected: number;
}

interface MinItemsErrorCause {
  details: MinItemsErrorCauseDetails;
  location: string;
  kind: "minItems";
}

interface MinimumErrorCauseDetails {
  actual: number;
  minimum: number;
}

interface MinimumErrorCause {
  details: MinimumErrorCauseDetails;
  location: string;
  kind: "minimum";
}

// union type of cause details, depend on kind
type ErrorCause =
  | RequiredErrorCause
  | GeneralErrorCause
  | FormatErrorCause
  | MinItemsErrorCause
  | MinimumErrorCause;

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

interface APIDuplicatedIdentityError {
  errorName: string;
  reason: "DuplicatedIdentity";
}

interface APIInvalidError {
  errorName: string;
  reason: "Invalid";
}

interface PasswordPolicyViolatedErrorCause {
  Name: string;
  Info: unknown;
}

interface PasswordPolicyViolatedErrorInfo {
  causes: PasswordPolicyViolatedErrorCause[];
}

interface APIPasswordPolicyViolatedError {
  errorName: string;
  info: PasswordPolicyViolatedErrorInfo;
  reason: "PasswordPolicyViolated";
}

// union type of api errors, depend on reason
type APIError =
  | APIValidationError
  | APIInvariantViolationError
  | APIInvalidError
  | APIDuplicatedIdentityError
  | APIPasswordPolicyViolatedError;

function isAPIError(value?: { [key: string]: any }): value is APIError {
  if (value == null) {
    return false;
  }
  return "errorName" in value && "reason" in value;
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
    case "format":
      return {
        kind: cause.kind,
        location: cause.location,
        detail: cause.details.format,
      };
    case "minItems":
      return {
        kind: cause.kind,
        location: cause.location,
        minItems: cause.details.expected,
      };
    case "minimum":
      return {
        kind: cause.kind,
        location: cause.location,
        minimum: cause.details.minimum,
      };
    default:
      return { kind: "Unknown" };
  }
}

export function handleUpdateAppConfigError(error: GraphQLError): Violation[] {
  const unknownViolation: Violation[] = [{ kind: "Unknown" }];
  if (!isAPIError(error.extensions)) {
    return unknownViolation;
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
    case "Invalid": {
      return [{ kind: "Invalid" }];
    }
    case "DuplicatedIdentity": {
      return [{ kind: "DuplicatedIdentity" }];
    }
    case "PasswordPolicyViolated": {
      const causes = extensions.info.causes;
      const causeNames = causes.map((cause) => cause.Name);
      return [{ kind: "PasswordPolicyViolated", causes: causeNames }];
    }
    default:
      return unknownViolation;
  }
}

export function parseError(error: unknown): Violation[] {
  if (error == null) {
    return [];
  }
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
  return [{ kind: "Unknown" }];
}
