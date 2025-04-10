import React from "react";
import cn from "classnames";
import { TextArea as RadixTextArea } from "@radix-ui/themes";
import styles from "./TextArea.module.css";
import { FormField } from "../FormField/FormField";

type TextAreaSize = "2" | "3";

export interface TextAreaProps {
  darkMode?: boolean;
  size: TextAreaSize;
  label?: React.ReactNode;
  optional?: boolean;
  disabled?: boolean;
  readOnly?: boolean;
  placeholder?: string;
  error?: React.ReactNode;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLTextAreaElement>;
}

export function TextArea({
  darkMode,
  size,
  label,
  optional,
  disabled,
  readOnly,
  placeholder,
  error,
  value,
  onChange,
}: TextAreaProps): React.ReactElement {
  return (
    <FormField
      darkMode={darkMode}
      size={size}
      label={label}
      optional={optional}
      error={error}
    >
      <RadixTextArea
        className={cn(error != null ? styles["textArea--error"] : null)}
        variant="surface"
        size={size}
        placeholder={placeholder}
        disabled={disabled}
        readOnly={readOnly}
        value={value}
        onChange={onChange}
      />
    </FormField>
  );
}
