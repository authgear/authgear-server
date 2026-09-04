import React, { useCallback } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { DateFieldDateTime } from "../v2/DateField/DateField";
import styles from "./FraudProtectionDateRangeDialog.module.css";

interface FraudProtectionDateRangeDialogProps {
  hidden: boolean;
  title: string;
  fromDatePickerLabel: string;
  toDatePickerLabel: string;
  rangeFrom?: Date;
  rangeTo?: Date;
  fromDatePickerMaxDate?: Date;
  toDatePickerMaxDate?: Date;
  onSelectRangeFrom?: (date: Date | null | undefined) => void;
  onSelectRangeTo?: (date: Date | null | undefined) => void;
  onCommitDateRange?: (e?: React.MouseEvent<unknown>) => void;
  onDismiss?: (e?: React.MouseEvent<unknown>) => void;
}

const FraudProtectionDateRangeDialog: React.VFC<FraudProtectionDateRangeDialogProps> =
  function FraudProtectionDateRangeDialog(props) {
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
        <Dialog.Content maxWidth="420px" size="3">
          <Dialog.Title>{title}</Dialog.Title>
          <div className={styles.fields}>
            <div className={styles.field}>
              <DateFieldDateTime
                size="2"
                label={fromDatePickerLabel}
                value={rangeFrom ?? null}
                max={toDatePickerMaxDate ?? fromDatePickerMaxDate}
                onChange={onChangeRangeFrom}
              />
            </div>
            <div className={styles.field}>
              <DateFieldDateTime
                size="2"
                label={toDatePickerLabel}
                value={rangeTo ?? null}
                max={toDatePickerMaxDate}
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

export default FraudProtectionDateRangeDialog;
