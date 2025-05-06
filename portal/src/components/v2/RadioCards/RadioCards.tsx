import cn from "classnames";
import { RadioCards as RadixRadioCards, Text } from "@radix-ui/themes";
import React, { useMemo } from "react";
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
  ...rootProps
}: RadioCardsProps<T>): React.ReactElement {
  const onToggleCallbacks = useMemo(() => {
    return options.map((option) => {
      const fn = () => {
        if (value === option.value) {
          return;
        }
        onValueChange(option.value);
      };
      return fn;
    });
  }, [onValueChange, options, value]);

  return (
    <Root {...rootProps}>
      {options.map((option, idx) => {
        return (
          <OptionItem
            key={option.value}
            option={option}
            checked={value === option.value}
            onToggle={onToggleCallbacks[idx]}
          />
        );
      })}
    </Root>
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
  ...rootProps
}: MultiSelectRadioCardsProps<T>): React.ReactElement {
  const checkedValuesSet = useMemo(() => new Set(values), [values]);

  const onToggleCallbacks = useMemo(() => {
    return options.map((option) => {
      const fn = () => {
        const newValues = new Set(checkedValuesSet);
        if (!checkedValuesSet.has(option.value)) {
          newValues.add(option.value);
        } else {
          newValues.delete(option.value);
        }
        onValuesChange(Array.from(newValues));
      };
      return fn;
    });
  }, [checkedValuesSet, onValuesChange, options]);

  return (
    <Root {...rootProps}>
      {options.map((option, idx) => {
        return (
          <OptionItem
            key={option.value}
            option={option}
            checked={checkedValuesSet.has(option.value)}
            onToggle={onToggleCallbacks[idx]}
          />
        );
      })}
    </Root>
  );
}

interface RootProps {
  darkMode?: boolean;
  highContrast?: boolean;
  size: "1" | "2" | "3";
  itemMinWidth?: number;
  itemFillSpaces?: boolean;
  numberOfColumns?: number;
  children?: React.ReactNode;
}

function Root({
  darkMode,
  size,
  highContrast,
  itemMinWidth = 160,
  itemFillSpaces = false,
  numberOfColumns,
  children,
}: RootProps) {
  return (
    <RadixRadioCards.Root
      className={cn(styles.radioCards__root, darkMode ? "dark" : null)}
      size={size}
      variant="surface"
      color="indigo"
      highContrast={highContrast}
      columns={`repeat(${gridColumnRepeat(
        numberOfColumns
      )}, minmax(${itemMinWidth}px, ${itemMaxSize(itemFillSpaces)}))`}
    >
      {children}
    </RadixRadioCards.Root>
  );
}

function OptionItem<T extends string>({
  option,
  checked,
  onToggle,
}: {
  option: RadioCardOption<T>;
  checked?: boolean;
  onToggle?: () => void;
}) {
  return (
    <RadixRadioCards.Item
      key={option.value}
      value={option.value}
      disabled={option.disabled}
      checked={checked}
      onClick={onToggle}
    >
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
    </RadixRadioCards.Item>
  );
}

function itemMaxSize(itemFillSpaces: boolean) {
  if (itemFillSpaces) {
    return "1fr";
  }
  return "max-content";
}

function gridColumnRepeat(numberOfColumns: number | undefined) {
  if (numberOfColumns == null) {
    return "auto-fit";
  }
  return `${numberOfColumns}`;
}
