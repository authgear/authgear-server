import React from "react";
import { Switch as RadixSwitch } from "@radix-ui/themes";
import styles from "./Toggle.module.css";

export interface ToggleProps {
  text?: React.ReactNode;
  disabled?: boolean;
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
}

export function Toggle({
  text,
  disabled,
  checked,
  onCheckedChange,
}: ToggleProps): React.ReactElement {
  return (
    <label className={styles.toggle}>
      <RadixSwitch
        disabled={disabled}
        checked={checked}
        onCheckedChange={onCheckedChange}
      />
      {text ? <p className={styles.toggle__text}>{text}</p> : null}
    </label>
  );
}
