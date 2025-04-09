import cn from "classnames";
import { RadioCards as RadixRadioCards } from "@radix-ui/themes";
import React from "react";
import styles from "./RadioCards.module.css";

export interface RadioCardOption<T extends string> {
  value: T;
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  disabled?: boolean;
}

export interface RadioCardsProps<T extends string> {
  darkMode?: boolean;
  highContrast?: boolean;
  size: "1" | "2" | "3";
  value: T | null;
  options: RadioCardOption<T>[];
  onValueChange: (newValue: T) => void;
}

export function RadioCards<T extends string>({
  darkMode,
  highContrast,
  size,
  value,
  onValueChange,
  options,
}: RadioCardsProps<T>): React.ReactElement {
  return (
    <RadixRadioCards.Root
      className={cn(styles.radioCards__root, darkMode ? "dark" : null)}
      size={size}
      variant="surface"
      color="indigo"
      highContrast={highContrast}
      value={value ?? undefined}
      onValueChange={onValueChange}
      columns="repeat(auto-fit, minmax(160px, max-content))"
    >
      {options.map((option) => {
        return (
          <RadixRadioCards.Item
            key={option.value}
            value={option.value}
            disabled={option.disabled}
          >
            <div className={styles.radioCards__itemTextContainer}>
              <p className={styles.radioCards__itemTextTitle}>{option.title}</p>
              {option.subtitle ? (
                <p className={styles.radioCards__itemTextSubtitle}>
                  {option.subtitle}
                </p>
              ) : null}
            </div>
          </RadixRadioCards.Item>
        );
      })}
    </RadixRadioCards.Root>
  );
}
