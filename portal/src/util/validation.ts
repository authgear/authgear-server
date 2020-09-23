import { Values } from "@oursky/react-messageformat";
import { nonNullable } from "./types";

// union type of different kind of violation
export type Violation =
  | RequiredViolation
  | GeneralViolation
  | RemoveLastIdentityViolation;

interface RequiredViolation {
  kind: "required";
  location: string;
  missingField: string[];
}

interface GeneralViolation {
  kind: "general";
  location: string;
}

interface RemoveLastIdentityViolation {
  kind: "RemoveLastIdentity";
}

// list of violation kind recognized
const violationKinds = ["required", "general", "RemoveLastIdentity"];
type ViolationKind = Violation["kind"];
export function isViolationKind(value?: string): value is ViolationKind {
  return value != null && violationKinds.includes(value);
}

type ViolationSelector = (violation: Violation) => boolean;
type ViolationSelectors<Key extends string> = Record<Key, ViolationSelector>;

function defaultFormatErrorMessageList(
  errorMessages: string[]
): string | undefined {
  return errorMessages.length === 0 ? undefined : errorMessages.join("\n");
}

// default seclectors
export function makeMissingFieldSelector(
  locationPrefix: string,
  fieldName: string
): ViolationSelector {
  return function (violation: Violation) {
    if (violation.kind !== "required") {
      return false;
    }
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
    default:
      return undefined;
  }
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
