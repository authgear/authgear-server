import React, { useMemo } from "react";
import cn from "classnames";
import { TextField as RadixTextField } from "@radix-ui/themes";
import styles from "./TextField.module.css";
import { InfoCircledIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { FormField } from "../FormField/FormField";
import { ErrorParseRule } from "../../../error/parse";
import { useErrorMessage } from "../../../formbinding";

type TextFieldSize = "2" | "3";

function Icon({ icon }: { icon: TextFieldIcon }): React.ReactElement {
  switch (icon) {
    case TextFieldIcon.MagnifyingGlass:
      return <MagnifyingGlassIcon className={styles.textField__icon} />;
    case TextFieldIcon.InfoCircled:
      return <InfoCircledIcon className={styles.textField__icon} />;
  }
}

export enum TextFieldIcon {
  MagnifyingGlass = "MagnifyingGlass",
  InfoCircled = "InfoCircled",
}

export interface TextInputProps {
  size: TextFieldSize;
  type?: "text" | "password" | "email" | "number" | "search" | "tel" | "url" | "hidden" | "date" | "time" | "datetime-local" | "month" | "week";
  disabled?: boolean;
  readOnly?: boolean;
  placeholder?: string;
  error?: React.ReactNode;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
  onFocus?: React.FocusEventHandler<HTMLInputElement>;
  onBlur?: React.FocusEventHandler<HTMLInputElement>;
}

export interface TextFieldProps extends TextInputProps {
  darkMode?: boolean;
  label?: React.ReactNode;
  /** Label typography size; defaults to `size` when omitted. */
  labelSize?: TextFieldSize;
  optional?: boolean;
  suffix?: React.ReactNode;
  /** Icon-only suffix (e.g. password visibility) without chip background/border. */
  suffixPlain?: boolean;
  iconStart?: TextFieldIcon;
  iconEnd?: TextFieldIcon;
  hint?: React.ReactNode;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;

  parentJSONPointer?: string | RegExp;
  fieldName?: string;
  errorRules?: ErrorParseRule[];
}

function TextField_(props: TextFieldProps): React.ReactElement {
  const {
    darkMode,
    size,
    label,
    labelSize,
    optional,
    error,
    hint,
    iconStart,
    iconEnd,
    suffix,
    suffixPlain,

    parentJSONPointer = "",
    fieldName,
    errorRules,
  } = props;
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

  return (
    <FormField
      darkMode={darkMode}
      size={size}
      labelSize={labelSize}
      label={label}
      optional={optional}
      error={error}
      hint={hint}
      labelSpace="1"
      parentJSONPointer={parentJSONPointer}
      fieldName={fieldName}
      errorRules={errorRules}
    >
      <Input
        {...props}
        disabled={props.disabled || fieldProps.disabled}
        error={props.error ?? fieldProps.errorMessage}
      >
        {iconStart != null ? (
          <RadixTextField.Slot side="left">
            <Icon icon={iconStart} />
          </RadixTextField.Slot>
        ) : null}
        {suffix != null ? (
          <RadixTextField.Slot
            className={cn(
              styles.textField__suffix,
              suffixPlain && styles["textField__suffix--plain"]
            )}
            side="right"
          >
            {suffix}
          </RadixTextField.Slot>
        ) : iconEnd != null ? (
          <RadixTextField.Slot side="right">
            <Icon icon={iconEnd} />
          </RadixTextField.Slot>
        ) : null}
      </Input>
    </FormField>
  );
}

function Input({
  size,
  type,
  disabled,
  readOnly,
  placeholder,
  error,
  value,
  onChange,
  onBlur,
  onFocus,
  children,
}: TextInputProps & { children: React.ReactNode }): React.ReactElement {
  return (
    <RadixTextField.Root
      className={cn(error != null ? styles["textField--error"] : null)}
      variant="surface"
      size={size}
      type={type}
      placeholder={placeholder}
      disabled={disabled}
      readOnly={readOnly}
      value={value}
      onChange={onChange}
      onBlur={onBlur}
      onFocus={onFocus}
    >
      {children}
    </RadixTextField.Root>
  );
}

export const TextField = Object.assign(TextField_, { Input });
