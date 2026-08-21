import React, { useMemo } from "react";
import cn from "classnames";
import { TextArea as RadixTextArea } from "@radix-ui/themes";
import styles from "./TextArea.module.css";
import { FormField } from "../FormField/FormField";
import { ErrorParseRule } from "../../../error/parse";
import { useErrorMessage } from "../../../formbinding";

type TextAreaSize = "2" | "3";

export interface TextAreaProps {
  darkMode?: boolean;
  size: TextAreaSize;
  label?: React.ReactNode;
  /** Label typography size; defaults to `size` when omitted. */
  labelSize?: TextAreaSize;
  optional?: boolean;
  required?: boolean;
  placeholder?: string;
  error?: React.ReactNode;
  hint?: React.ReactNode;
  className?: string;
  disabled?: boolean;
  readOnly?: boolean;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLTextAreaElement>;

  parentJSONPointer?: string | RegExp;
  fieldName?: string;
  errorRules?: ErrorParseRule[];
}

export function TextArea({
  darkMode,
  size,
  label,
  labelSize,
  optional,
  required,
  placeholder,
  error,
  hint,
  className,
  value,
  onChange,
  parentJSONPointer = "",
  fieldName,
  errorRules,
  disabled,
  readOnly,
}: TextAreaProps): React.ReactElement {
  const field = useMemo(
    () =>
      fieldName != null
        ? {
            parentJSONPointer,
            fieldName,
            rules: errorRules,
          }
        : undefined,
    [parentJSONPointer, fieldName, errorRules]
  );

  const fieldProps = useErrorMessage(field);
  const hasError = error != null || fieldProps.errorMessage != null;

  return (
    <FormField
      darkMode={darkMode}
      size={size}
      labelSize={labelSize}
      label={label}
      optional={optional}
      required={required}
      error={error ?? fieldProps.errorMessage}
      hint={hint}
      labelSpace="1"
    >
      <RadixTextArea
        className={cn(hasError ? styles["textArea--error"] : null, className)}
        variant="surface"
        size={size}
        placeholder={placeholder}
        disabled={(disabled ?? false) || fieldProps.disabled}
        readOnly={readOnly}
        value={value}
        onChange={onChange}
      />
    </FormField>
  );
}
