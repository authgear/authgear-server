import React, { useCallback } from "react";
import { Dialog } from "@radix-ui/themes";
import { DateTime } from "luxon";
import { FormattedMessage } from "../../intl";
import styles from "./DateRangeDialog.module.css";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { DateField } from "../../components/v2/DateField/DateField";
import DateTimePicker from "../../DateTimePicker";

interface DateRangeDialogBaseProps {
  hidden: boolean;
  title: string;
  fromDatePickerLabel: string;
  toDatePickerLabel: string;
  rangeFrom?: Date;
  rangeTo?: Date;
  onSelectRangeFrom?: (date: Date | null | undefined) => void;
  onSelectRangeTo?: (date: Date | null | undefined) => void;
  onCommitDateRange?: (e?: React.MouseEvent<unknown>) => void;
  onDismiss?: (e?: React.MouseEvent<unknown>) => void;
}

// DateTimePicker only supports "now" as its lower bound (see DateTimePickerProps.minDateTime),
// so fromDatePickerMinDate/toDatePickerMinDate have no equivalent when showTimePicker is true.
// Splitting the props by showTimePicker prevents them from being passed (and silently ignored)
// together.
type DateRangeDialogProps = DateRangeDialogBaseProps &
  (
    | {
        showTimePicker: true;
        fromDatePickerMaxDate?: Date;
        toDatePickerMaxDate?: Date;
      }
    | {
        showTimePicker?: false;
        fromDatePickerMinDate?: Date;
        fromDatePickerMaxDate?: Date;
        toDatePickerMinDate?: Date;
        toDatePickerMaxDate?: Date;
      }
  );

function toFieldValue(date?: Date): string {
  return date != null ? DateTime.fromJSDate(date).toFormat("yyyy-LL-dd") : "";
}

function fromFieldValue(value: string): Date | null {
  // The native date input only ever emits "" or a valid yyyy-MM-dd string.
  if (value === "") {
    return null;
  }
  return DateTime.fromISO(value).toJSDate();
}

const DateRangeDialog: React.VFC<DateRangeDialogProps> =
  function DateRangeDialog(props) {
    const {
      hidden,
      title,
      fromDatePickerLabel,
      toDatePickerLabel,
      rangeFrom,
      rangeTo,
      fromDatePickerMaxDate,
      toDatePickerMaxDate,
      onSelectRangeFrom,
      onSelectRangeTo,
      onCommitDateRange,
      onDismiss,
      showTimePicker = false,
    } = props;
    const fromDatePickerMinDate = props.showTimePicker
      ? undefined
      : props.fromDatePickerMinDate;
    const toDatePickerMinDate = props.showTimePicker
      ? undefined
      : props.toDatePickerMinDate;

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss?.();
        }
      },
      [onDismiss]
    );

    const onFromFieldChange = useCallback(
      (value: string) => {
        onSelectRangeFrom?.(fromFieldValue(value));
      },
      [onSelectRangeFrom]
    );

    const onToFieldChange = useCallback(
      (value: string) => {
        onSelectRangeTo?.(fromFieldValue(value));
      },
      [onSelectRangeTo]
    );

    return (
      <Dialog.Root open={!hidden} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth={showTimePicker ? "480px" : "340px"} size="3">
          <Dialog.Title>{title}</Dialog.Title>
          <div className={styles.fields}>
            {showTimePicker ? (
              <>
                <DateTimePicker
                  className={styles.dateTimePicker}
                  label={
                    <span className={styles.dateTimePickerLabel}>
                      {fromDatePickerLabel}
                    </span>
                  }
                  pickedDateTime={rangeFrom ?? null}
                  minDateTime={null}
                  maxDateTime={fromDatePickerMaxDate ?? null}
                  onPickDateTime={onSelectRangeFrom ?? (() => {})}
                  showClearButton={false}
                />
                <DateTimePicker
                  className={styles.dateTimePicker}
                  label={
                    <span className={styles.dateTimePickerLabel}>
                      {toDatePickerLabel}
                    </span>
                  }
                  pickedDateTime={rangeTo ?? null}
                  minDateTime={null}
                  maxDateTime={toDatePickerMaxDate ?? null}
                  onPickDateTime={onSelectRangeTo ?? (() => {})}
                  showClearButton={false}
                />
              </>
            ) : (
              <>
                <DateField
                  size="2"
                  label={fromDatePickerLabel}
                  value={toFieldValue(rangeFrom)}
                  min={
                    fromDatePickerMinDate != null
                      ? toFieldValue(fromDatePickerMinDate)
                      : undefined
                  }
                  max={
                    fromDatePickerMaxDate != null
                      ? toFieldValue(fromDatePickerMaxDate)
                      : undefined
                  }
                  onChange={onFromFieldChange}
                />
                <DateField
                  size="2"
                  label={toDatePickerLabel}
                  value={toFieldValue(rangeTo)}
                  min={
                    toDatePickerMinDate != null
                      ? toFieldValue(toDatePickerMinDate)
                      : undefined
                  }
                  max={
                    toDatePickerMaxDate != null
                      ? toFieldValue(toDatePickerMaxDate)
                      : undefined
                  }
                  onChange={onToFieldChange}
                />
              </>
            )}
          </div>
          <div className={styles.actions}>
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
          </div>
        </Dialog.Content>
      </Dialog.Root>
    );
  };

export default DateRangeDialog;
