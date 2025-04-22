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

export interface TextFieldProps {
  darkMode?: boolean;
  size: TextFieldSize;
  label?: React.ReactNode;
  optional?: boolean;
  disabled?: boolean;
  readOnly?: boolean;
  placeholder?: string;
  error?: React.ReactNode;
  suffix?: React.ReactNode;
  hint?: React.ReactNode;
  iconStart?: TextFieldIcon;
  iconEnd?: TextFieldIcon;

  value?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
}

export function TextField({
  darkMode,
  size,
  label,
  optional,
  disabled,
  readOnly,
  placeholder,
  error,
  hint,
  suffix,
  iconStart,
  iconEnd,
  value,
  onChange,
}: TextFieldProps): React.ReactElement {
  return (
    <FormField
      darkMode={darkMode}
      size={size}
      label={label}
      optional={optional}
      error={error}
      hint={hint}
    >
      <RadixTextField.Root
        className={cn(error != null ? styles["textField--error"] : null)}
        variant="surface"
        size={size}
        placeholder={placeholder}
        disabled={disabled}
        readOnly={readOnly}
        value={value}
        onChange={onChange}
      >
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
      </RadixTextField.Root>
    </FormField>
  );
}
