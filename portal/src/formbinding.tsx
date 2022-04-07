/* global JSX */
// This file contains the binding between our form abstraction and fluent ui.
import React, { useMemo, useContext } from "react";
import { Context, Values } from "@oursky/react-messageformat";
import { useFormField } from "./form";
import { FormField, ParsedAPIError } from "./error/parse";
import ErrorRenderer from "./ErrorRenderer";

export interface FieldProps<T> {
  disabled: boolean;
  errorMessage?: T;
}

export function useErrorMessage(formField: FormField): FieldProps<JSX.Element> {
  const { loading, errors } = useFormField(formField);
  if (errors.length <= 0) {
    return {
      disabled: loading,
    };
  }
  return {
    disabled: loading,
    errorMessage: <ErrorRenderer errors={errors} />,
  };
}

function renderErrors(
  errors: readonly ParsedAPIError[],
  renderMessage: (id: string, args: Values) => string
): string | undefined {
  if (errors.length === 0) {
    return undefined;
  }
  return errors.map((err) => renderError(err, renderMessage)).join("\n");
}

function renderError(
  error: ParsedAPIError,
  renderMessage: (id: string, args: Values) => string
): string {
  if (error.messageID) {
    const args: Values = { ...error.arguments };
    return renderMessage(error.messageID, args);
  }
  return error.message ?? "";
}

export function useErrorMessageString(
  formField: FormField
): FieldProps<string> {
  const { renderToString } = useContext(Context);
  const { loading, errors } = useFormField(formField);
  const errorMessage = useMemo(
    () => renderErrors(errors, renderToString),
    [errors, renderToString]
  );
  return {
    disabled: loading,
    errorMessage,
  };
}
