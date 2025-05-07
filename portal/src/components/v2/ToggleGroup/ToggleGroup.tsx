import React, { useCallback, useMemo } from "react";
import { Text } from "@radix-ui/themes";
import cn from "classnames";

import { Toggle } from "../Toggle/Toggle";
import styles from "./ToggleGroup.module.css";

export interface ToggleGroupOption<T extends string> {
  value: T;
  text: React.ReactNode;
  icon?: React.ReactNode;
  supportingText?: React.ReactNode;
}

export interface ToggleGroupProps<T extends string> {
  className?: string;
  items: ToggleGroupOption<T>[];
  values: T[];
  onValuesChange?: (newValues: T[]) => void;
  onToggle?: (checked: boolean, value: T) => void;
}

export function ToggleGroup<T extends string>({
  className,
  items,
  values,
  onValuesChange,
  onToggle,
}: ToggleGroupProps<T>): React.ReactElement {
  const valuesSet = useMemo(() => new Set(values), [values]);

  const handleCheckedChange = useCallback(
    (checked: boolean, value: T) => {
      onToggle?.(checked, value);
      const newValuesSet = new Set(values);
      if (checked) {
        newValuesSet.add(value);
      } else {
        newValuesSet.delete(value);
      }
      onValuesChange?.(Array.from(newValuesSet));
    },
    [onValuesChange, onToggle, values]
  );

  return (
    <div className={cn(styles.toggleGroup, className)}>
      {items.map((it, idx) => (
        <React.Fragment key={it.value}>
          <ToggleGroupItem
            value={it.value}
            text={it.text}
            icon={it.icon}
            supportingText={it.supportingText}
            checked={valuesSet.has(it.value)}
            onCheckedChange={handleCheckedChange}
          />
          {idx !== items.length - 1 ? (
            <hr className={styles.toggleGroup__itemSeparator} key={it.value} />
          ) : null}
        </React.Fragment>
      ))}
    </div>
  );
}

export interface ToggleGroupItemProps<T extends string>
  extends ToggleGroupOption<T> {
  checked?: boolean;
  onCheckedChange?: (checked: boolean, value: T) => void;
}

export function ToggleGroupItem<T extends string>({
  value,
  text,
  icon,
  supportingText,
  checked,
  onCheckedChange,
}: ToggleGroupItemProps<T>): React.ReactElement {
  return (
    <label className={styles.toggleGroupItem}>
      <div className={styles.toggleGroupItem__leftColumn}>
        {icon ? icon : null}
        <div className={styles.toggleGroupItem__textContainer}>
          <Text
            as="p"
            size={"3"}
            weight={"medium"}
            className={styles.toggleGroupItem__text}
          >
            {text}
          </Text>
          {supportingText ? (
            <Text
              as="p"
              size={"2"}
              weight={"regular"}
              className={styles.toggleGroupItem__supportingText}
            >
              {supportingText}
            </Text>
          ) : null}
        </div>
      </div>
      <Toggle
        checked={checked}
        onCheckedChange={useCallback(
          (checked: boolean) => {
            onCheckedChange?.(checked, value);
          },
          [onCheckedChange, value]
        )}
      />
    </label>
  );
}
