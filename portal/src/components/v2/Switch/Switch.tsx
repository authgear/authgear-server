import React from "react";
import { Switch as RadixSwitch } from "@radix-ui/themes";
import styles from "./Switch.module.css";

export interface SwitchProps {
  text?: React.ReactNode;
  disabled?: boolean;
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
}

export function Switch({
  text,
  disabled,
  checked,
  onCheckedChange,
}: SwitchProps): React.ReactElement {
  return (
    <label className={styles.switch}>
      <RadixSwitch
        disabled={disabled}
        checked={checked}
        onCheckedChange={onCheckedChange}
      />
      {text ? <p className={styles.switch__text}>{text}</p> : null}
    </label>
  );
}
