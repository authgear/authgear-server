// This file contains the binding between our form abstraction and fluent ui.
import { useMemo, useContext } from "react";
import { Context } from "@oursky/react-messageformat";
import { useFormField } from "./form";
import { FormField, renderErrors } from "./error/parse";

export interface FieldProps {
  errorMessage?: string;
}

export function useErrorMessage(formField: FormField): FieldProps {
  const { renderToString } = useContext(Context);
  const { errors } = useFormField(formField);
  const errorMessage = useMemo(
    () => renderErrors(errors, renderToString),
    [errors, renderToString]
  );
  return {
    errorMessage,
  };
}
