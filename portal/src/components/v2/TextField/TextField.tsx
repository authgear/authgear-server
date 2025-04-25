import React from "react";
import cn from "classnames";
import { TextField as RadixTextField } from "@radix-ui/themes";
import styles from "./TextField.module.css";
import { InfoCircledIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { FormField } from "../FormField/FormField";

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
  disabled?: boolean;
  readOnly?: boolean;
  placeholder?: string;
  error?: React.ReactNode;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
  onBlur?: React.FocusEventHandler<HTMLInputElement>;
}

export interface TextFieldProps extends TextInputProps {
  darkMode?: boolean;
  label?: React.ReactNode;
  optional?: boolean;
  suffix?: React.ReactNode;
  iconStart?: TextFieldIcon;
  iconEnd?: TextFieldIcon;
  hint?: React.ReactNode;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
}

function TextField_(props: TextFieldProps): React.ReactElement {
  const {
    darkMode,
    size,
    label,
    optional,
    error,
    hint,
    iconStart,
    iconEnd,
    suffix,
  } = props;
  return (
    <FormField
      darkMode={darkMode}
      size={size}
      label={label}
      optional={optional}
      error={error}
      hint={hint}
      labelSpace="1"
    >
      <Input {...props}>
        {iconStart != null ? (
          <RadixTextField.Slot side="left">
            <Icon icon={iconStart} />
          </RadixTextField.Slot>
        ) : null}
        {suffix != null ? (
          <RadixTextField.Slot
            className={styles.textField__suffix}
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
  disabled,
  readOnly,
  placeholder,
  error,
  value,
  onChange,
  onBlur,
  children,
}: TextInputProps & { children: React.ReactNode }): React.ReactElement {
  return (
    <RadixTextField.Root
      className={cn(error != null ? styles["textField--error"] : null)}
      variant="surface"
      size={size}
      placeholder={placeholder}
      disabled={disabled}
      readOnly={readOnly}
      value={value}
      onChange={onChange}
      onBlur={onBlur}
    >
      {children}
    </RadixTextField.Root>
  );
}

export const TextField = Object.assign(TextField_, { Input });
