import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { CalendarIcon, ChevronDownIcon } from "@radix-ui/react-icons";
import { Popover, RadioGroup } from "@radix-ui/themes";
import { Context as MessageContext } from "../../intl";
import {
  AUDIT_LOG_DATE_RANGE_PRESET_ORDER,
  AuditLogDateRangePresetKey,
  formatCustomDateRangeLabel,
  formatDateRangeBoundLabel,
} from "./dateRangePresets";
import styles from "./AuditLogDateRangeFilterDropdown.module.css";

export type { AuditLogDateRangePresetKey };

interface AuditLogDateRangeFilterDropdownProps {
  className?: string;
  value: AuditLogDateRangePresetKey;
  onChange: (value: AuditLogDateRangePresetKey) => void;
  rangeFrom?: Date | null;
  rangeTo?: Date | null;
  // Include the time of day in the custom range label when a bound is not on
  // a day boundary. Leave off for date-only pickers, whose Dates may carry an
  // arbitrary time (e.g. UTC midnight).
  showTime?: boolean;
  onOpenCustomDateRangeDialog?: () => void;
  presets?: AuditLogDateRangePresetKey[];
}

export const AuditLogDateRangeFilterDropdown: React.VFC<AuditLogDateRangeFilterDropdownProps> =
  function AuditLogDateRangeFilterDropdown({
    className,
    value,
    onChange,
    rangeFrom = null,
    rangeTo = null,
    showTime = false,
    onOpenCustomDateRangeDialog,
    presets = AUDIT_LOG_DATE_RANGE_PRESET_ORDER,
  }) {
    const { renderToString, locale } = useContext(MessageContext);
    const [open, setOpen] = useState(false);

    const optionLabels = useMemo(() => {
      return {
        today: renderToString("AuditLogScreen.date-range.today"),
        last7Days: renderToString("AuditLogScreen.date-range.last-7-days"),
        last30Days: renderToString("AuditLogScreen.date-range.last-30-days"),
        custom: renderToString("AuditLogScreen.date-range.custom"),
      };
    }, [renderToString]);

    const customRangeLabel = useMemo(() => {
      if (rangeFrom != null && rangeTo != null) {
        return formatCustomDateRangeLabel(locale, rangeFrom, rangeTo, showTime);
      }
      // Open-ended range: only one bound was set.
      if (rangeFrom != null) {
        return renderToString("AuditLogScreen.date-range.from-only", {
          from: formatDateRangeBoundLabel(locale, rangeFrom, "from", showTime),
        });
      }
      if (rangeTo != null) {
        return renderToString("AuditLogScreen.date-range.to-only", {
          to: formatDateRangeBoundLabel(locale, rangeTo, "to", showTime),
        });
      }
      return null;
    }, [locale, rangeFrom, rangeTo, renderToString, showTime]);

    const customDateInputPlaceholder = renderToString(
      "AuditLogScreen.date-range.custom-placeholder"
    );

    const selectedLabel = useMemo(() => {
      if (value === "custom" && customRangeLabel != null) {
        return customRangeLabel;
      }
      return optionLabels[value];
    }, [customRangeLabel, optionLabels, value]);

    const showCustomRangeLabel = value === "custom" && customRangeLabel != null;
    const showCustomDateInput = value === "custom";

    const onSelectOption = useCallback(
      (nextValue: AuditLogDateRangePresetKey) => {
        onChange(nextValue);
        if (nextValue !== "custom") {
          setOpen(false);
        }
      },
      [onChange]
    );

    const onCustomDateInputClick = useCallback(
      (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setOpen(false);
        onOpenCustomDateRangeDialog?.();
      },
      [onOpenCustomDateRangeDialog]
    );

    const onOptionClick = useCallback(
      (optionKey: AuditLogDateRangePresetKey, e: React.MouseEvent) => {
        if (optionKey === "custom" && value === "custom") {
          onCustomDateInputClick(e);
        }
      },
      [onCustomDateInputClick, value]
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
              title={showCustomRangeLabel ? selectedLabel : undefined}
            >
              <span className={styles.triggerLabel}>{selectedLabel}</span>
              <ChevronDownIcon className={styles.triggerIcon} />
            </button>
          </Popover.Trigger>
          <Popover.Content
            className={cn(
              styles.content,
              showCustomDateInput && styles.contentWithCustom
            )}
            sideOffset={4}
            align="start"
          >
            <RadioGroup.Root
              value={value}
              onValueChange={(nextValue) => {
                onSelectOption(nextValue as AuditLogDateRangePresetKey);
              }}
            >
              {presets.map((optionKey) => (
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
                  {optionKey === "custom" && showCustomDateInput ? (
                    <div className={styles.customDateInputWrapper}>
                      <button
                        type="button"
                        className={styles.customDateInput}
                        aria-label={customDateInputPlaceholder}
                        onClick={onCustomDateInputClick}
                      >
                        <span
                          className={cn(
                            styles.customDateInputLabel,
                            customRangeLabel == null &&
                              styles.customDateInputLabelPlaceholder
                          )}
                        >
                          {customRangeLabel ?? customDateInputPlaceholder}
                        </span>
                        <CalendarIcon className={styles.customDateInputIcon} />
                      </button>
                    </div>
                  ) : null}
                </div>
              ))}
            </RadioGroup.Root>
          </Popover.Content>
        </Popover.Root>
      </div>
    );
  };
