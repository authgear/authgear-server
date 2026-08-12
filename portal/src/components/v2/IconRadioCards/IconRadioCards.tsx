import cn from "classnames";
import {
  CheckboxCards as RadixCheckboxCards,
  RadioCards as RadixRadioCards,
  Text,
} from "@radix-ui/themes";
import React, { useCallback } from "react";
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
  // RadixRadioCards.Item derives its checked visual from
  // Root's value === Item's value, so the Root must be controlled for the
  // form state to fully drive the selection.
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
    <Root {...rootProps} value={value ?? ""} onValueChange={handleValueChange}>
      {options.map((option) => {
        return <OptionItem key={option.value} option={option} />;
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
  size,
  itemMinWidth = 160,
  itemFillSpaces = false,
  numberOfColumns,
}: MultiSelectIconRadioCardsProps<T>): React.ReactElement {
  // A radio group can only ever hold one value, so the multi-select variant
  // is backed by CheckboxCards; its Root is controlled by the whole value
  // list, which also keeps a forced-on disabled item rendered as checked.
  const handleValueChange = useCallback(
    (newValues: string[]) => {
      onValuesChange(newValues as T[]);
    },
    [onValuesChange]
  );

  return (
    <RadixCheckboxCards.Root
      className={cn(styles.iconRadioCards__root)}
      size={size}
      variant="surface"
      color="indigo"
      value={values}
      onValueChange={handleValueChange}
      columns={`repeat(${gridColumnRepeat(
        numberOfColumns
      )}, minmax(${itemMinWidth}px, ${itemMaxSize(itemFillSpaces)}))`}
    >
      {options.map((option) => {
        return (
          <Tooltip
            key={option.value}
            content={option.tooltip}
            disabled={option.tooltip == null}
          >
            {/* We need this extra div because Tooltip and RadixCheckboxCards.Item both write to data-state attribute causing bugs */}
            <div className={styles.iconRadioCards__itemWrapper}>
              <RadixCheckboxCards.Item
                className={styles.iconRadioCards__item}
                value={option.value}
                disabled={option.disabled}
              >
                <OptionItemContent option={option} />
              </RadixCheckboxCards.Item>
            </div>
          </Tooltip>
        );
      })}
    </RadixCheckboxCards.Root>
  );
}

interface RootProps {
  size: "2" | "3";
  itemMinWidth?: number;
  itemFillSpaces?: boolean;
  numberOfColumns?: number;
  value?: string;
  onValueChange?: (newValue: string) => void;
  children?: React.ReactNode;
}

function Root({
  size,
  itemMinWidth = 160,
  itemFillSpaces = false,
  numberOfColumns,
  value,
  onValueChange,
  children,
}: RootProps) {
  return (
    <RadixRadioCards.Root
      className={cn(styles.iconRadioCards__root)}
      size={size}
      variant="surface"
      color="indigo"
      value={value}
      onValueChange={onValueChange}
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
}: {
  option: IconRadioCardOption<T>;
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
        >
          <OptionItemContent option={option} />
        </RadixRadioCards.Item>
      </div>
    </Tooltip>
  );
}

function OptionItemContent<T extends string>({
  option,
}: {
  option: IconRadioCardOption<T>;
}) {
  return (
    <div
      className={cn(
        styles.iconRadioCards__itemContainer,
        option.subtitle == null &&
          styles["iconRadioCards__itemContainer--center"]
      )}
    >
      <div className={styles.iconRadioCards__iconContainer}>{option.icon}</div>
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
