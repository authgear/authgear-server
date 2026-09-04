import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { ChevronDownIcon } from "@radix-ui/react-icons";
import { Popover, RadioGroup } from "@radix-ui/themes";
import { Context as MessageContext } from "../../intl";
import styles from "./AuditLogDateRangeFilterDropdown.module.css";

export type DateRangeFilterDropdownOptionKey =
  | "allDateRange"
  | "customDateRange";

interface DateRangeFilterDropdownProps {
  className?: string;
  value: DateRangeFilterDropdownOptionKey;
  customRangeLabel?: string;
  onClickAllDateRange: (
    e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>
  ) => void;
  onClickCustomDateRange: (
    e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>
  ) => void;
}

export const DateRangeFilterDropdown: React.VFC<DateRangeFilterDropdownProps> =
  function DateRangeFilterDropdown({
    className,
    value,
    customRangeLabel,
    onClickAllDateRange,
    onClickCustomDateRange,
  }: DateRangeFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);
    const [open, setOpen] = useState(false);

    const optionLabels = useMemo(() => {
      return {
        allDateRange: renderToString("AuditLogScreen.date-range.all"),
        customDateRange: renderToString("AuditLogScreen.date-range.custom"),
      };
    }, [renderToString]);

    const showCustomRangeLabel =
      value === "customDateRange" && customRangeLabel != null;
    const selectedLabel = showCustomRangeLabel
      ? customRangeLabel
      : optionLabels[value];

    const onSelectOption = useCallback(
      (nextValue: DateRangeFilterDropdownOptionKey) => {
        setOpen(false);
        if (nextValue === "allDateRange") {
          onClickAllDateRange();
        } else {
          onClickCustomDateRange();
        }
      },
      [onClickAllDateRange, onClickCustomDateRange]
    );

    const onOptionClick = useCallback(
      (optionKey: DateRangeFilterDropdownOptionKey, e: React.MouseEvent) => {
        // Reopen the custom range dialog when the already selected custom
        // option is clicked again.
        if (optionKey === "customDateRange" && value === "customDateRange") {
          e.preventDefault();
          e.stopPropagation();
          setOpen(false);
          onClickCustomDateRange();
        }
      },
      [onClickCustomDateRange, value]
    );

    const onOpenChange = useCallback((nextOpen: boolean) => {
      setOpen(nextOpen);
    }, []);

    return (
      <div
        className={cn(
          styles.root,
          showCustomRangeLabel && styles.rootCustom,
          className
        )}
      >
        <Popover.Root open={open} onOpenChange={onOpenChange}>
          <Popover.Trigger>
            <button
              type="button"
              className={styles.trigger}
              aria-label={selectedLabel}
            >
              <span className={styles.triggerLabel}>{selectedLabel}</span>
              <ChevronDownIcon className={styles.triggerIcon} />
            </button>
          </Popover.Trigger>
          <Popover.Content
            className={styles.content}
            sideOffset={4}
            align="start"
          >
            <RadioGroup.Root
              value={value}
              onValueChange={(nextValue) => {
                onSelectOption(nextValue as DateRangeFilterDropdownOptionKey);
              }}
            >
              {(
                [
                  "allDateRange",
                  "customDateRange",
                ] as DateRangeFilterDropdownOptionKey[]
              ).map((optionKey) => (
                <div key={optionKey} className={styles.optionGroup}>
                  <label
                    className={styles.option}
                    onClick={(e) => {
                      onOptionClick(optionKey, e);
                    }}
                  >
                    <RadioGroup.Item value={optionKey} />
                    <span className={styles.optionLabel}>
                      {optionLabels[optionKey]}
                    </span>
                  </label>
                </div>
              ))}
            </RadioGroup.Root>
          </Popover.Content>
        </Popover.Root>
      </div>
    );
  };
