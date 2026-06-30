import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { ChevronDownIcon } from "@radix-ui/react-icons";
import { Popover, RadioGroup } from "@radix-ui/themes";
import { Context as MessageContext } from "../../intl";
import {
  AUDIT_LOG_DATE_RANGE_PRESET_ORDER,
  AuditLogDateRangePresetKey,
} from "./dateRangePresets";
import styles from "./AuditLogDateRangeFilterDropdown.module.css";

export type { AuditLogDateRangePresetKey };

interface AuditLogDateRangeFilterDropdownProps {
  className?: string;
  value: AuditLogDateRangePresetKey;
  onChange: (value: AuditLogDateRangePresetKey) => void;
}

export const AuditLogDateRangeFilterDropdown: React.VFC<AuditLogDateRangeFilterDropdownProps> =
  function AuditLogDateRangeFilterDropdown({ className, value, onChange }) {
    const { renderToString } = useContext(MessageContext);
    const [open, setOpen] = useState(false);

    const optionLabels = useMemo(() => {
      return {
        today: renderToString("AuditLogScreen.date-range.today"),
        last7Days: renderToString("AuditLogScreen.date-range.last-7-days"),
        last30Days: renderToString("AuditLogScreen.date-range.last-30-days"),
        custom: renderToString("AuditLogScreen.date-range.custom"),
      };
    }, [renderToString]);

    const selectedLabel = optionLabels[value];

    const onSelectOption = useCallback(
      (nextValue: AuditLogDateRangePresetKey) => {
        onChange(nextValue);
        setOpen(false);
      },
      [onChange]
    );

    const onOpenChange = useCallback((nextOpen: boolean) => {
      setOpen(nextOpen);
    }, []);

    return (
      <div className={cn(styles.root, className)}>
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
                onSelectOption(nextValue as AuditLogDateRangePresetKey);
              }}
            >
              {AUDIT_LOG_DATE_RANGE_PRESET_ORDER.map((optionKey) => (
                <label key={optionKey} className={styles.option}>
                  <RadioGroup.Item value={optionKey} />
                  <span className={styles.optionLabel}>
                    {optionLabels[optionKey]}
                  </span>
                </label>
              ))}
            </RadioGroup.Root>
          </Popover.Content>
        </Popover.Root>
      </div>
    );
  };
