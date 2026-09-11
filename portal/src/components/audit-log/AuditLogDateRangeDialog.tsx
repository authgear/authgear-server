import React, { useCallback, useMemo } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import {
  DateFieldDate,
  DateFieldDateTime,
  toDateString,
} from "../v2/DateField/DateField";
import styles from "./AuditLogDateRangeDialog.module.css";

interface AuditLogDateRangeDialogProps {
  hidden: boolean;
  title: string;
  fromDatePickerLabel: string;
  toDatePickerLabel: string;
  rangeFrom?: Date;
  rangeTo?: Date;
  fromDatePickerMinDate?: Date;
  fromDatePickerMaxDate?: Date;
  toDatePickerMinDate?: Date;
  toDatePickerMaxDate?: Date;
  // Pick a time of day as well as a date. The min/max bounds then apply as
  // instants rather than as calendar days.
  showTimePicker?: boolean;
  onSelectRangeFrom?: (date: Date | null | undefined) => void;
  onSelectRangeTo?: (date: Date | null | undefined) => void;
  onCommitDateRange?: (e?: React.MouseEvent<unknown>) => void;
  onDismiss?: (e?: React.MouseEvent<unknown>) => void;
}

function toMinMaxString(date: Date | undefined): string | undefined {
  if (date == null) {
    return undefined;
  }
  return toDateString(date);
}

interface RangeBoundFieldProps {
  showTimePicker: boolean;
  label: string;
  value: Date | undefined;
  minDate: Date | undefined;
  maxDate: Date | undefined;
  onChange: (date: Date | null) => void;
}

function RangeBoundField(props: RangeBoundFieldProps): React.ReactElement {
  const { showTimePicker, label, value, minDate, maxDate, onChange } = props;

  const min = useMemo(() => toMinMaxString(minDate), [minDate]);
  const max = useMemo(() => toMinMaxString(maxDate), [maxDate]);

  if (showTimePicker) {
    return (
      <DateFieldDateTime
        size="2"
        label={label}
        value={value ?? null}
        min={minDate}
        max={maxDate}
        onChange={onChange}
      />
    );
  }
  return (
    <DateFieldDate
      size="2"
      label={label}
      value={value ?? null}
      min={min}
      max={max}
      onChange={onChange}
    />
  );
}

const AuditLogDateRangeDialog: React.VFC<AuditLogDateRangeDialogProps> =
  function AuditLogDateRangeDialog(props) {
    const {
      hidden,
      title,
      fromDatePickerLabel,
      toDatePickerLabel,
      rangeFrom,
      rangeTo,
      fromDatePickerMinDate,
      fromDatePickerMaxDate,
      toDatePickerMinDate,
      toDatePickerMaxDate,
      showTimePicker = false,
      onSelectRangeFrom,
      onSelectRangeTo,
      onCommitDateRange,
      onDismiss,
    } = props;

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss?.();
        }
      },
      [onDismiss]
    );

    const onChangeRangeFrom = useCallback(
      (date: Date | null) => {
        onSelectRangeFrom?.(date);
      },
      [onSelectRangeFrom]
    );

    const onChangeRangeTo = useCallback(
      (date: Date | null) => {
        onSelectRangeTo?.(date);
      },
      [onSelectRangeTo]
    );

    return (
      <Dialog.Root open={!hidden} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth={showTimePicker ? "420px" : "400px"} size="3">
          <Dialog.Title>{title}</Dialog.Title>
          <div className={styles.fields}>
            <div className={styles.field}>
              <RangeBoundField
                showTimePicker={showTimePicker}
                label={fromDatePickerLabel}
                value={rangeFrom}
                minDate={fromDatePickerMinDate}
                maxDate={fromDatePickerMaxDate}
                onChange={onChangeRangeFrom}
              />
            </div>
            <div className={styles.field}>
              <RangeBoundField
                showTimePicker={showTimePicker}
                label={toDatePickerLabel}
                value={rangeTo}
                minDate={toDatePickerMinDate}
                maxDate={toDatePickerMaxDate}
                onChange={onChangeRangeTo}
              />
            </div>
          </div>
          <Flex gap="3" mt="4" justify="end">
            <SecondaryButton
              size="2"
              onClick={onDismiss}
              text={<FormattedMessage id="cancel" />}
            />
            <PrimaryButton
              size="2"
              onClick={onCommitDateRange}
              text={<FormattedMessage id="done" />}
            />
          </Flex>
        </Dialog.Content>
      </Dialog.Root>
    );
  };

export default AuditLogDateRangeDialog;
