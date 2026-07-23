import React, { useMemo } from "react";
import cn from "classnames";
import styles from "./FormField.module.css";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../../intl";
import { ErrorParseRule } from "../../../error/parse";
import { useErrorMessage } from "../../../formbinding";

type FormFieldSize = "2" | "3";

type FormFieldLabelSpace = "1" | "2";

export interface FormFieldProps {
  darkMode?: boolean;
  size: FormFieldSize;
  /** Label typography size; defaults to `size` when omitted. */
  labelSize?: FormFieldSize;
  label?: React.ReactNode;
  /** Associates the label with the control it wraps (renders the label as a `<label htmlFor>`). */
  htmlFor?: string;
  optional?: boolean;
  required?: boolean;
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
  labelSize,
  label,
  htmlFor,
  optional,
  required,
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

  const labelContent = label ? (
    <>
      {label}
      {required ? (
        <span className={styles.formField__labelRequired} aria-hidden="true">
          &nbsp;*
        </span>
      ) : null}
      {optional ? (
        <span className={styles.formField__labelOptional}>
          &nbsp;
          <FormattedMessage id="FormField.optional" />
        </span>
      ) : null}
    </>
  ) : null;

  return (
    <div
      className={cn(
        styles.formField,
        labelSpaceClassName(labelSpace),
        darkMode ? "dark" : null
      )}
    >
      {labelContent != null ? (
        htmlFor != null ? (
          <Text
            as="label"
            htmlFor={htmlFor}
            size={labelSize ?? size}
            weight={"medium"}
            className={styles.formField__label}
          >
            {labelContent}
          </Text>
        ) : (
          <Text
            as="p"
            size={labelSize ?? size}
            weight={"medium"}
            className={styles.formField__label}
          >
            {labelContent}
          </Text>
        )
      ) : null}
      <div className={styles.formField__inputContainer}>
        {children}
        {error != null ? (
          <Text
            as="p"
            className={styles.formField__errorMessage}
            size={"1"}
            weight={"regular"}
          >
            {error}
          </Text>
        ) : null}
        {hint != null ? (
          <Text
            as="p"
            className={styles.formField__hint}
            size={"1"}
            weight={"regular"}
          >
            {hint}
          </Text>
        ) : null}
      </div>
    </div>
  );
}

function labelSpaceClassName(space: FormFieldLabelSpace) {
  switch (space) {
    case "1":
      return styles["formField--space1"];
    case "2":
      return styles["formField--space2"];
  }
}
