import React, { useCallback, useMemo } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { DateFieldDate, toDateString } from "../v2/DateField/DateField";
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

    const fromMin = useMemo(
      () => toMinMaxString(fromDatePickerMinDate),
      [fromDatePickerMinDate]
    );
    const fromMax = useMemo(
      () => toMinMaxString(fromDatePickerMaxDate),
      [fromDatePickerMaxDate]
    );
    const toMin = useMemo(
      () => toMinMaxString(toDatePickerMinDate),
      [toDatePickerMinDate]
    );
    const toMax = useMemo(
      () => toMinMaxString(toDatePickerMaxDate),
      [toDatePickerMaxDate]
    );

    return (
      <Dialog.Root open={!hidden} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>{title}</Dialog.Title>
          <div className={styles.fields}>
            <div className={styles.field}>
              <DateFieldDate
                size="2"
                label={fromDatePickerLabel}
                value={rangeFrom ?? null}
                min={fromMin}
                max={fromMax}
                onChange={onChangeRangeFrom}
              />
            </div>
            <div className={styles.field}>
              <DateFieldDate
                size="2"
                label={toDatePickerLabel}
                value={rangeTo ?? null}
                min={toMin}
                max={toMax}
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
