import cn from "classnames";
import {
  CheckboxCards as RadixCheckboxCards,
  RadioCards as RadixRadioCards,
  Text,
} from "@radix-ui/themes";
import React, { useCallback } from "react";
import styles from "./RadioCards.module.css";

export interface RadioCardOption<T extends string> {
  value: T;
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  disabled?: boolean;
}

interface RadioCardsPropsBase<T extends string> {
  darkMode?: boolean;
  highContrast?: boolean;
  size: "1" | "2" | "3";
  options: RadioCardOption<T>[];
  itemMinWidth?: number;
  itemFillSpaces?: boolean;
  numberOfColumns?: number;
}

export interface RadioCardsProps<T extends string>
  extends RadioCardsPropsBase<T> {
  value: T | null;
  onValueChange: (newValue: T) => void;
}

export function RadioCards<T extends string>({
  value,
  onValueChange,
  options,
  darkMode,
  size,
  highContrast,
  itemMinWidth,
  itemFillSpaces,
  numberOfColumns,
}: RadioCardsProps<T>): React.ReactElement {
  // RadixRadioCards.Item derives its checked visual solely from
  // Root's value === Item's value; a `checked` prop on the Item is ignored.
  // The Root must therefore be controlled for the form state to drive the
  // selection (e.g. restoring a persisted selection on mount).
  const handleValueChange = useCallback(
    (newValue: string) => {
      if (newValue === value) {
        return;
      }
      onValueChange(newValue as T);
    },
    [onValueChange, value]
  );

  return (
    <RadixRadioCards.Root
      className={cn(styles.radioCards__root, darkMode ? "dark" : null)}
      size={size}
      variant="surface"
      color="indigo"
      highContrast={highContrast}
      columns={gridColumns(numberOfColumns, itemMinWidth, itemFillSpaces)}
      // "" (never undefined) keeps the Root controlled from the first render,
      // so a null selection renders as nothing checked instead of leaving the
      // group uncontrolled. Matches IconRadioCards.
      value={value ?? ""}
      onValueChange={handleValueChange}
    >
      {options.map((option) => {
        return (
          <RadixRadioCards.Item
            key={option.value}
            value={option.value}
            disabled={option.disabled}
          >
            <OptionItemContent option={option} />
          </RadixRadioCards.Item>
        );
      })}
    </RadixRadioCards.Root>
  );
}

export interface MultiSelectRadioCardsProps<T extends string>
  extends RadioCardsPropsBase<T> {
  values: T[];
  onValuesChange: (newValues: T[]) => void;
}

export function MultiSelectRadioCards<T extends string>({
  values,
  onValuesChange,
  options,
  darkMode,
  size,
  highContrast,
  itemMinWidth,
  itemFillSpaces,
  numberOfColumns,
}: MultiSelectRadioCardsProps<T>): React.ReactElement {
  // A radio group can only ever hold one value, so multi-select must be
  // backed by CheckboxCards; its Root is controlled by the whole value list.
  const handleValueChange = useCallback(
    (newValues: string[]) => {
      onValuesChange(newValues as T[]);
    },
    [onValuesChange]
  );

  return (
    <RadixCheckboxCards.Root
      className={cn(styles.radioCards__root, darkMode ? "dark" : null)}
      size={size}
      variant="surface"
      color="indigo"
      highContrast={highContrast}
      columns={gridColumns(numberOfColumns, itemMinWidth, itemFillSpaces)}
      value={values}
      onValueChange={handleValueChange}
    >
      {options.map((option) => {
        return (
          <RadixCheckboxCards.Item
            key={option.value}
            value={option.value}
            disabled={option.disabled}
          >
            <OptionItemContent option={option} />
          </RadixCheckboxCards.Item>
        );
      })}
    </RadixCheckboxCards.Root>
  );
}

function OptionItemContent<T extends string>({
  option,
}: {
  option: RadioCardOption<T>;
}) {
  return (
    <div className={styles.radioCards__itemTextContainer}>
      <Text
        as="p"
        size={"2"}
        weight={"medium"}
        className={styles.radioCards__itemTextTitle}
      >
        {option.title}
      </Text>
      {option.subtitle ? (
        <Text
          as="p"
          size={"2"}
          weight={"regular"}
          className={styles.radioCards__itemTextSubtitle}
        >
          {option.subtitle}
        </Text>
      ) : null}
    </div>
  );
}

function gridColumns(
  numberOfColumns: number | undefined,
  itemMinWidth: number = 160,
  itemFillSpaces: boolean = false
) {
  const repeat = numberOfColumns == null ? "auto-fit" : `${numberOfColumns}`;
  const maxSize = itemFillSpaces ? "1fr" : "max-content";
  return `repeat(${repeat}, minmax(${itemMinWidth}px, ${maxSize}))`;
}
