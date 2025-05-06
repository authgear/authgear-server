import cn from "classnames";
import { RadioCards as RadixRadioCards, Text } from "@radix-ui/themes";
import React, { useMemo } from "react";
import styles from "./IconRadioCards.module.css";
import { Tooltip } from "../Tooltip/Tooltip";

export interface IconRadioCardOption<T extends string> {
  value: T;
  icon: React.ReactNode;
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  tooltip?: React.ReactNode;
  disabled?: boolean;
}

interface IconRadioCardsPropsBase<T extends string> {
  size: "2" | "3";
  options: IconRadioCardOption<T>[];
  itemMinWidth?: number;
  itemFillSpaces?: boolean;
  numberOfColumns?: number;
}

export interface IconRadioCardsProps<T extends string>
  extends IconRadioCardsPropsBase<T> {
  value: T | null;
  onValueChange: (newValue: T) => void;
}

export function IconRadioCards<T extends string>({
  value,
  onValueChange,
  options,
  ...rootProps
}: IconRadioCardsProps<T>): React.ReactElement {
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

export interface MultiSelectIconRadioCardsProps<T extends string>
  extends IconRadioCardsPropsBase<T> {
  values: T[];
  onValuesChange: (newValues: T[]) => void;
}

export function MultiSelectIconRadioCards<T extends string>({
  values,
  onValuesChange,
  options,
  ...rootProps
}: MultiSelectIconRadioCardsProps<T>): React.ReactElement {
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
  size: "2" | "3";
  itemMinWidth?: number;
  itemFillSpaces?: boolean;
  numberOfColumns?: number;
  children?: React.ReactNode;
}

function Root({
  size,
  itemMinWidth = 160,
  itemFillSpaces = false,
  numberOfColumns,
  children,
}: RootProps) {
  return (
    <RadixRadioCards.Root
      className={cn(styles.iconRadioCards__root)}
      size={size}
      variant="surface"
      color="indigo"
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
  option: IconRadioCardOption<T>;
  checked?: boolean;
  onToggle?: () => void;
}) {
  return (
    <Tooltip content={option.tooltip} disabled={option.tooltip == null}>
      {/* We need this extra div because Tooltip and RadixRadioCards.Item both write to data-state attribute causing bugs */}
      <div className={styles.iconRadioCards__itemWrapper}>
        <RadixRadioCards.Item
          className={styles.iconRadioCards__item}
          key={option.value}
          value={option.value}
          disabled={option.disabled}
          checked={checked}
          onClick={onToggle}
        >
          <div className={styles.iconRadioCards__itemContainer}>
            <div className={styles.iconRadioCards__iconContainer}>
              {option.icon}
            </div>
            <div className={styles.iconRadioCards__itemTextContainer}>
              <Text
                as="p"
                size={"2"}
                weight={"medium"}
                className={styles.iconRadioCards__itemTextTitle}
              >
                {option.title}
              </Text>
              {option.subtitle ? (
                <Text
                  as="p"
                  size={"2"}
                  weight={"regular"}
                  className={styles.iconRadioCards__itemTextSubtitle}
                >
                  {option.subtitle}
                </Text>
              ) : null}
            </div>
          </div>
        </RadixRadioCards.Item>
      </div>
    </Tooltip>
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
