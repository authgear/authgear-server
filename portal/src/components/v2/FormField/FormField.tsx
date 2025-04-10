import React from "react";
import cn from "classnames";
import styles from "./FormField.module.css";

type FormFieldSize = "2" | "3";

export interface FormFieldProps {
  darkMode?: boolean;
  size: FormFieldSize;
  label?: React.ReactNode;
  error?: React.ReactNode;
  children?: React.ReactNode;
}

export function FormField({
  darkMode,
  size,
  label,
  error,
  children,
}: FormFieldProps): React.ReactElement {
  return (
    <label className={cn(styles.formField, darkMode ? "dark" : null)}>
      {label ? (
        <p className={cn(styles.formField__label, sizeToLabelClass(size))}>
          {label}
        </p>
      ) : null}
      {children}
      {error != null ? (
        <p className={styles.textField__errorMessage}>{error}</p>
      ) : null}
    </label>
  );
}

function sizeToLabelClass(size: FormFieldSize) {
  switch (size) {
    case "2":
      return styles["formField__label--size2"];
    case "3":
      return styles["formField__label--size3"];
  }
}
