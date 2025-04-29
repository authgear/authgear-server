import React, { useMemo } from "react";
import cn from "classnames";
import styles from "./FormField.module.css";
import { FormattedMessage } from "@oursky/react-messageformat";
import { ErrorParseRule } from "../../../error/parse";
import { useErrorMessage } from "../../../formbinding";

type FormFieldSize = "2" | "3";

type FormFieldLabelSpace = "1" | "2";

export interface FormFieldProps {
  darkMode?: boolean;
  size: FormFieldSize;
  label?: React.ReactNode;
  optional?: boolean;
  error?: React.ReactNode;
  hint?: React.ReactNode;
  children?: React.ReactNode;

  labelSpace?: FormFieldLabelSpace;

  parentJSONPointer?: string | RegExp;
  fieldName?: string;
  errorRules?: ErrorParseRule[];
}

export function FormField({
  darkMode,
  size,
  label,
  optional,
  error: propsError,
  hint,
  children,

  labelSpace = "2",

  parentJSONPointer = "",
  fieldName,
  errorRules,
}: FormFieldProps): React.ReactElement {
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

  const error = propsError ?? fieldProps.errorMessage;

  return (
    <div
      className={cn(
        styles.formField,
        labelSpaceClassName(labelSpace),
        darkMode ? "dark" : null
      )}
    >
      {label ? (
        <p className={cn(styles.formField__label, sizeToLabelClass(size))}>
          {label}
          {optional ? (
            <span className={styles.formField__labelOptional}>
              &nbsp;
              <FormattedMessage id="FormField.optional" />
            </span>
          ) : null}
        </p>
      ) : null}
      <div className={styles.formField__inputContainer}>
        {children}
        {error != null ? (
          <p className={styles.formField__errorMessage}>{error}</p>
        ) : null}
        {hint != null ? <p className={styles.formField__hint}>{hint}</p> : null}
      </div>
    </div>
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

function labelSpaceClassName(space: FormFieldLabelSpace) {
  switch (space) {
    case "1":
      return styles["formField--space1"];
    case "2":
      return styles["formField--space2"];
  }
}
