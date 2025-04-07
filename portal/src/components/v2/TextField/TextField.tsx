import React from "react";
import cn from "classnames";
import { TextField as RadixTextField } from "@radix-ui/themes";
import styles from "./TextField.module.css";

type TextFieldSize = "2" | "3";

function sizeToLabelClass(size: TextFieldSize) {
  switch (size) {
    case "2":
      return styles["textField__label--size2"];
    case "3":
      return styles["textField__label--size3"];
  }
}

export interface TextFieldProps {
  darkMode?: boolean;
  size: TextFieldSize;
  label?: React.ReactNode;
  disabled?: boolean;
  readOnly?: boolean;
  placeholder?: string;
  error?: React.ReactNode;
}

export function TextField({
  darkMode,
  size,
  label,
  disabled,
  readOnly,
  placeholder,
  error,
}: TextFieldProps): React.ReactElement {
  return (
    <label className={cn(styles.textField, darkMode ? "dark" : null)}>
      {label ? (
        <p className={cn(styles.textField__label, sizeToLabelClass(size))}>
          {label}
        </p>
      ) : null}
      <RadixTextField.Root
        className={cn(error != null ? styles["textField--error"] : null)}
        variant="surface"
        size={size}
        placeholder={placeholder}
        disabled={disabled}
        readOnly={readOnly}
      />
      {error != null ? (
        <p className={styles.textField__errorMessage}>{error}</p>
      ) : null}
    </label>
  );
}
