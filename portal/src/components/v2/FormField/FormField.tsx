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
        <Text
          as="p"
          size={size}
          weight={"medium"}
          className={styles.formField__label}
        >
          {label}
          {optional ? (
            <span className={styles.formField__labelOptional}>
              &nbsp;
              <FormattedMessage id="FormField.optional" />
            </span>
          ) : null}
        </Text>
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
