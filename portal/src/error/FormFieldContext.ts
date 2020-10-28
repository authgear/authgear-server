import { Context, Values } from "@oursky/react-messageformat";
import { useContext, useEffect } from "react";

import { useFormContext } from "./FormContext";
import { ValidationFailedErrorInfoCause } from "./validation";

const errorMessageID: Record<ValidationFailedErrorInfoCause["kind"], string> = {
  required: "validation-error.required",
  general: "validation-error.general",
  format: "validation-error.format",
  minItems: "validation-error.minItems",
  minimum: "validation-error.minimum",
  maximum: "validation-error.maximum",
  minLength: "validation-error.minLength",
  maxLength: "validation-error.maxLength",
};

interface FormFieldData {
  errorMessage: string | undefined;
}

function getReactMessageFormatValues(
  cause: ValidationFailedErrorInfoCause
): Values {
  return cause.details;
}

function constructErrorMessageFromValidationErrorCause(
  renderToString: (messageID: string, values?: Values) => string,
  cause: ValidationFailedErrorInfoCause,
  fieldName: string,
  fieldNameMessageID?: string
): string | undefined {
  // special handle required violation, needs to match missing field
  if (cause.kind === "required") {
    if (cause.details.missing.includes(fieldName)) {
      // fallback to raw field name if field name message ID not exist
      let localizedFieldName = fieldName;
      if (fieldNameMessageID != null) {
        localizedFieldName = renderToString(fieldNameMessageID);
      } else {
        console.warn(
          "[Construct validation error message]: Expect fieldNameMessageID in rules for `required` violation error"
        );
      }
      return renderToString(errorMessageID["required"], {
        fieldName: localizedFieldName,
      });
    }
  }
  // other than required violation, matching json pointer => matching field
  // unrecognized error kind (violation type) => throw error in get message value
  try {
    const errorMessageValues = getReactMessageFormatValues(cause);
    const messageID = errorMessageID[cause.kind];
    // NOTE: catch error cause not defined in type definition
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    if (messageID == null) {
      throw new Error();
    }
    return renderToString(messageID, errorMessageValues);
  } catch {
    console.warn(
      "[Unhandled validation error cause]: Unrecognized cause kind\n",
      cause
    );
    return undefined;
  }
}

export function useFormField(
  jsonPointer: RegExp | string,
  parentJSONPointer: RegExp | string,
  fieldName: string,
  fieldNameMessageID?: string
): FormFieldData {
  const { causes: formErrorCauses, registerField } = useFormContext();
  const { renderToString } = useContext(Context);
  // register field
  useEffect(() => {
    registerField(jsonPointer, parentJSONPointer, fieldName);
  }, [registerField, fieldName, jsonPointer, parentJSONPointer]);

  // handle error cause
  const fieldErrorCauses = formErrorCauses[String(jsonPointer)] ?? [];
  const errorMessageList: string[] = [];
  for (const cause of fieldErrorCauses) {
    const errorMessage = constructErrorMessageFromValidationErrorCause(
      renderToString,
      cause,
      fieldName,
      fieldNameMessageID
    );
    if (errorMessage != null) {
      errorMessageList.push(errorMessage);
    } else {
      console.warn("[Form field]: Unrecognized error cause");
    }
  }
  const fieldErrorMessage = errorMessageList.join("\n");

  return { errorMessage: fieldErrorMessage };
}
