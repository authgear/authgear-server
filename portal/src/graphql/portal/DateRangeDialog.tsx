import { DatePicker, Dialog, DialogFooter } from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import React, { useMemo } from "react";
import styles from "./DateRangeDialog.module.css";
import TextField from "../../TextField";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

interface DateRangeDialogProps {
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
  onSelectRangeFrom?: (date: Date | null | undefined) => void;
  onSelectRangeTo?: (date: Date | null | undefined) => void;
  onCommitDateRange?: (e?: React.MouseEvent<unknown>) => void;
  onDismiss?: (e?: React.MouseEvent<unknown>) => void;
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
      fromDatePickerMinDate,
      fromDatePickerMaxDate,
      toDatePickerMinDate,
      toDatePickerMaxDate,
      onSelectRangeFrom,
      onSelectRangeTo,
      onCommitDateRange,
      onDismiss,
    } = props;

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
        /* https://developer.microsoft.com/en-us/fluentui#/controls/web/dialog
         * Best practice says the max width is 340 */
        minWidth={340}
      >
        {/* Dialog is based on Modal, which will focus the first child on open. *
    However, we do not want the date picker to be opened at the same time. *
    So we make the first focusable element a hidden TextField */}
        <TextField className={styles.hidden} />
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
