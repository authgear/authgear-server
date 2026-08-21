import React from "react";
import { Switch as RadixSwitch, Text } from "@radix-ui/themes";
import styles from "./Toggle.module.css";

export interface ToggleProps {
  text?: React.ReactNode;
  textWeight?: "regular" | "medium" | "bold";
  disabled?: boolean;
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
}

export function Toggle({
  text,
  textWeight = "regular",
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
      {text ? (
        <Text
          as="p"
          size={"2"}
          weight={textWeight}
          className={styles.toggle__text}
        >
          {text}
        </Text>
      ) : null}
    </label>
  );
}
