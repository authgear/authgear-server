import { DatePicker, Dialog, DialogFooter } from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import React, { useMemo } from "react";
import styles from "./DateRangeDialog.module.css";
import TextField from "../../TextField";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
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

    const dateRangeDialogContentProps = useMemo(() => {
      return {
        title,
      };
    }, [title]);

    return (
      <Dialog
        hidden={hidden}
        onDismiss={onDismiss}
        dialogContentProps={dateRangeDialogContentProps}
        minWidth={showTimePicker ? 480 : 340}
      >
        {/* Dialog is based on Modal, which will focus the first child on open. *
    However, we do not want the date picker to be opened at the same time. *
    So we make the first focusable element a hidden TextField */}
        <TextField className={styles.hidden} />
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
              defaultTime={rangeFrom ?? null}
              defaultTimeOfDay="endOfDay"
              onPickDateTime={onSelectRangeTo ?? (() => {})}
              showClearButton={false}
            />
          </>
        ) : (
          <>
            <DatePicker
              label={fromDatePickerLabel}
              value={rangeFrom}
              minDate={fromDatePickerMinDate}
              maxDate={fromDatePickerMaxDate}
              onSelectDate={onSelectRangeFrom}
            />
            <DatePicker
              label={toDatePickerLabel}
              value={rangeTo}
              minDate={toDatePickerMinDate}
              maxDate={toDatePickerMaxDate}
              onSelectDate={onSelectRangeTo}
            />
          </>
        )}
        <DialogFooter>
          <PrimaryButton
            onClick={onCommitDateRange}
            text={<FormattedMessage id="done" />}
          />
          <DefaultButton
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };

export default DateRangeDialog;
