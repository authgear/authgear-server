import React, { useCallback, useContext } from "react";
import { DatePicker } from "@fluentui/react";
import { Dialog, Flex, Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { formatDateOnly } from "../../util/formatDateOnly";
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

    const { locale } = useContext(MessageContext);

    const formatDate = useCallback(
      (date?: Date) => {
        if (date == null) {
          return "";
        }
        return formatDateOnly(locale, date) ?? "";
      },
      [locale]
    );

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss?.();
        }
      },
      [onDismiss]
    );

    // The FluentUI DatePicker renders its calendar in a portaled FluentUI Layer
    // outside the Radix dialog. Without this guard, interacting with the calendar
    // counts as an "outside" interaction and closes the dialog.
    const onInteractOutside = useCallback((e: Event) => {
      const target = e.target as HTMLElement | null;
      if (target?.closest(".ms-Layer, .ms-Callout, .ms-DatePicker-callout")) {
        e.preventDefault();
      }
    }, []);

    return (
      <Dialog.Root open={!hidden} onOpenChange={onOpenChange}>
        <Dialog.Content
          maxWidth="400px"
          size="3"
          onInteractOutside={onInteractOutside}
        >
          <Dialog.Title>{title}</Dialog.Title>
          <div className={styles.fields}>
            <div className={styles.field}>
              <Text as="label" size="2" weight="medium">
                {fromDatePickerLabel}
              </Text>
              <DatePicker
                value={rangeFrom}
                minDate={fromDatePickerMinDate}
                maxDate={fromDatePickerMaxDate}
                formatDate={formatDate}
                onSelectDate={onSelectRangeFrom}
              />
            </div>
            <div className={styles.field}>
              <Text as="label" size="2" weight="medium">
                {toDatePickerLabel}
              </Text>
              <DatePicker
                value={rangeTo}
                minDate={toDatePickerMinDate}
                maxDate={toDatePickerMaxDate}
                formatDate={formatDate}
                onSelectDate={onSelectRangeTo}
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
