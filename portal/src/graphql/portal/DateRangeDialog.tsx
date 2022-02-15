import {
  DatePicker,
  DefaultButton,
  Dialog,
  DialogFooter,
  PrimaryButton,
  TextField,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React, { useMemo } from "react";
import styles from "./DateRangeDialog.module.scss";

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

const DateRangeDialog: React.FC<DateRangeDialogProps> =
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
          <PrimaryButton onClick={onCommitDateRange}>
            <FormattedMessage id="done" />
          </PrimaryButton>
          <DefaultButton onClick={onDismiss}>
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
    );
  };

export default DateRangeDialog;
