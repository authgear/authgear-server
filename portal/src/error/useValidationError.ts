import { useCallback, useState } from "react";

import { FormContextValue } from "./FormContext";
import { extractAPIError, isApolloError } from "./error";
import { ValidationFailedErrorInfoCause } from "./validation";

interface FieldRegister {
  jsonPointer: RegExp | string;
  parentJSONPointer: RegExp | string;
  fieldName: string;
}

type ErrorCausesMap = Partial<Record<string, ValidationFailedErrorInfoCause[]>>;

export function isLocationMatchWithJSONPointer(
  jsonPointer: RegExp | string,
  location: string
): boolean {
  if (typeof jsonPointer === "string") {
    return location === jsonPointer;
  }
  return jsonPointer.test(location);
}

function commonSelector(
  cause: ValidationFailedErrorInfoCause,
  jsonPointer: RegExp | string
): boolean {
  return isLocationMatchWithJSONPointer(jsonPointer, cause.location);
}

function requiredCauseSelector(
  cause: ValidationFailedErrorInfoCause,
  parentJSONPointer: string | RegExp,
  fieldName: string
) {
  if (cause.kind !== "required") {
    return false;
  }
  if (!isLocationMatchWithJSONPointer(parentJSONPointer, cause.location)) {
    return false;
  }
  return cause.details.missing.includes(fieldName);
}

function handleValidationErrorCause(
  cause: ValidationFailedErrorInfoCause,
  fields: FieldRegister[],
  matchedCauses: ErrorCausesMap,
  unhandledCauses: ValidationFailedErrorInfoCause[]
) {
  for (const field of fields) {
    const isMatch =
      requiredCauseSelector(cause, field.parentJSONPointer, field.fieldName) ||
      commonSelector(cause, field.jsonPointer);
    const jsonPointerString = String(field.jsonPointer);
    if (isMatch) {
      matchedCauses[jsonPointerString] = matchedCauses[jsonPointerString] ?? [];
      matchedCauses[jsonPointerString]?.push(cause);
      return;
    }
  }

  // no matching fields
  unhandledCauses.push(cause);
}

export function useValidationError(
  error: unknown
): {
  unhandledCauses?: ValidationFailedErrorInfoCause[];
  otherError?: unknown;
  value: FormContextValue;
} {
  const [fields, setFields] = useState<FieldRegister[]>([]);
  const registerField = useCallback(
    (
      jsonPointer: RegExp | string,
      parentJSONPointer: RegExp | string,
      fieldName: string
    ) => {
      setFields((prev) => [
        ...prev,
        { jsonPointer, parentJSONPointer, fieldName },
      ]);
    },
    []
  );
  const unhandledCauses: ValidationFailedErrorInfoCause[] = [];
  const matchedCauses: ErrorCausesMap = {};

  if (error == null) {
    return { unhandledCauses, value: { registerField, causes: {} } };
  }
  if (!isApolloError(error)) {
    return { otherError: error, value: { registerField, causes: {} } };
  }
  const { graphQLErrors } = error;
  for (const graphQLError of graphQLErrors) {
    const apiError = extractAPIError(graphQLError);
    if (apiError == null || apiError.reason !== "ValidationFailed") {
      return { otherError: error, value: { registerField, causes: {} } };
    }
    const { causes } = apiError.info;
    for (const cause of causes) {
      handleValidationErrorCause(cause, fields, matchedCauses, unhandledCauses);
    }
  }
  return {
    unhandledCauses,
    value: { registerField, causes: matchedCauses },
    otherError: unhandledCauses.length > 0 ? error : undefined,
  };
}
