/* global JSX */
// This file contains the binding between our form abstraction and fluent ui.
import React, {
  useMemo,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";
import { IDialogProps } from "@fluentui/react";
import { Context, Values } from "./intl";
import { useFormField } from "./form";
import { FormField, ParsedAPIError } from "./error/parse";
import ErrorRenderer from "./ErrorRenderer";

export interface FieldProps<T> {
  disabled: boolean;
  errorMessage?: T;
}

export interface DialogProps {
  hidden: IDialogProps["hidden"];
  onDismiss: IDialogProps["onDismiss"];
  errors: readonly ParsedAPIError[];
}

export function useErrorDialog(formField: FormField): DialogProps {
  const [hidden, setHidden] = useState(true);
  const { errors } = useFormField(formField);

  const onDismiss = useCallback((e) => {
    e?.preventDefault();
    e?.stopPropagation();
    setHidden(true);
  }, []);

  useEffect(() => {
    if (errors.length > 0) {
      setHidden(false);
    }
  }, [errors.length]);

  return {
    hidden,
    onDismiss,
    errors,
  };
}

export function useErrorMessage(
  formField: FormField | undefined
): FieldProps<JSX.Element> {
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
