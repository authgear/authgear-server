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
      return <MagnifyingGlassIcon className={styles.textFieldIcon} />;
    case TextFieldIcon.InfoCircled:
      return <InfoCircledIcon className={styles.textFieldIcon} />;
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
  disabled?: boolean;
  readOnly?: boolean;
  placeholder?: string;
  error?: React.ReactNode;
  iconStart?: TextFieldIcon;
  iconEnd?: TextFieldIcon;
}

export function TextField({
  darkMode,
  size,
  label,
  disabled,
  readOnly,
  placeholder,
  error,
  iconStart,
  iconEnd,
}: TextFieldProps): React.ReactElement {
  return (
    <FormField darkMode={darkMode} size={size} label={label} error={error}>
      <RadixTextField.Root
        className={cn(error != null ? styles["textField--error"] : null)}
        variant="surface"
        size={size}
        placeholder={placeholder}
        disabled={disabled}
        readOnly={readOnly}
      >
        <RadixTextField.Slot>
          {iconStart != null ? <Icon icon={iconStart} /> : null}
        </RadixTextField.Slot>
        <RadixTextField.Slot>
          {iconEnd != null ? <Icon icon={iconEnd} /> : null}
        </RadixTextField.Slot>
      </RadixTextField.Root>
    </FormField>
  );
}
