import { ApolloError } from "@apollo/client";
import { GraphQLError } from "graphql";
import { nonNullable } from "./types";
import { Values } from "@oursky/react-messageformat";

// union type of different kind of violation
export type Violation = RequiredViolation;

interface RequiredViolation {
  kind: "required";
  location: string;
  missingField: string[];
}

// list of violation kind recognized
const violationKinds = ["required"];
type ViolationKind = Violation["kind"];
function isViolationKind(value?: string): value is ViolationKind {
  return value != null && violationKinds.includes(value);
}

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

// union type of cause details, depend on kind
type ErrorCause = RequiredErrorCause;

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

type ViolationSelector = (violation: Violation) => boolean;
type ViolationSelectors<Key extends string> = Record<Key, ViolationSelector>;

function isAPIError(value?: { [key: string]: any }): value is APIError {
  if (value == null) {
    return false;
  }
  return "errorName" in value && "info" in value && "reason" in value;
}

function defaultFormatErrorMessageList(
  errorMessages: string[]
): string | undefined {
  return errorMessages.length === 0 ? undefined : errorMessages.join("\n");
}

function extractViolationFromErrorCause(cause: ErrorCause): Violation | null {
  if (!isViolationKind(cause.kind)) {
    return null;
  }
  return {
    kind: cause.kind,
    missingField: cause.details.missing,
    location: cause.location,
  };
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

// default seclectors
export function makeMissingFieldSelector(
  locationPrefix: string,
  fieldName: string
): ViolationSelector {
  return function (violation: Violation) {
    /* uncomment when there is more than one violation kind
    if (violation.kind !== "required") {
      return false
    }
    */
    if (!violation.location.startsWith(locationPrefix)) {
      return false;
    }
    return violation.missingField.includes(fieldName);
  };
}

export function combineSelector(
  selectorList: ViolationSelector[]
): ViolationSelector {
  return function (violation: Violation) {
    for (const selector of selectorList) {
      if (selector(violation)) {
        return true;
      }
    }
    return false;
  };
}

export function violationSelector<Key extends string>(
  violations: Violation[],
  violationSelectors: ViolationSelectors<Key>
): Record<Key, Violation[]> {
  const violationMap = Object.entries(violationSelectors).reduce<
    Partial<Record<Key, Violation[]>>
  >((violationMap, [key, selector]) => {
    violationMap[key as Key] = violations.filter(selector as ViolationSelector);
    return violationMap;
  }, {});
  return violationMap as Record<Key, Violation[]>;
}

function violationFormatter(
  fieldNameId: string,
  violation: Violation,
  renderToString: (messageId: string, values?: Values) => string
): string | undefined {
  switch (violation.kind) {
    case "required":
      return renderToString("required-field-missing", {
        fieldName: renderToString(fieldNameId),
      });
  }
  return undefined;
}

export function errorFormatter(
  fieldNameId: string,
  violations: Violation[],
  renderToString: (messageId: string, values?: Values) => string,
  formatErrorMessageList: (
    errorMessages: string[]
  ) => string | undefined = defaultFormatErrorMessageList
): string | undefined {
  return formatErrorMessageList(
    violations
      .map((violation) =>
        violationFormatter(fieldNameId, violation, renderToString)
      )
      .filter(nonNullable)
  );
}
