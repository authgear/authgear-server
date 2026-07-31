import React, { useId, useMemo } from "react";
import { Callout, Text } from "@radix-ui/themes";
import { ExclamationTriangleIcon, InfoCircledIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../intl";
import { useErrorMessage } from "../../formbinding";
import { TextField } from "../v2/TextField/TextField";
import { FormField } from "../v2/FormField/FormField";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { IPCheckResult } from "./IPBlocklistForm";
import styles from "./IPBlocklistCheckIPPanel.module.css";

export interface IPBlocklistCheckIPPanelProps {
  ipToCheck: string;
  onIPToCheckChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onCheckIP: () => void;
  checkingIP: boolean;
  checkIPResult: IPCheckResult | null;
}

export function IPBlocklistCheckIPPanel({
  ipToCheck,
  onIPToCheckChange,
  onCheckIP,
  checkingIP,
  checkIPResult,
}: IPBlocklistCheckIPPanelProps): React.ReactElement {
  const id = useId();
  const field = useMemo(
    () => ({
      parentJSONPointer: "",
      fieldName: "ipAddress",
    }),
    []
  );
  const fieldProps = useErrorMessage(field);

  return (
    <div className={styles.root}>
      <FormField
        size="2"
        labelSize="2"
        label={
          <FormattedMessage id="IPBlocklistForm.check-ip-address.label" />
        }
        htmlFor={id}
        labelSpace="1"
        parentJSONPointer=""
        fieldName="ipAddress"
      >
        <div className={styles.inputRow}>
          <TextField.Input
            id={id}
            size="2"
            inputClassName={styles.input}
            value={ipToCheck}
            onChange={onIPToCheckChange}
            disabled={fieldProps.disabled}
            error={fieldProps.errorMessage}
          >
            {null}
          </TextField.Input>
          <Text as="p" size="1" className={styles.hint}>
            <FormattedMessage id="IPBlocklistForm.check-ip-address.description" />
          </Text>
          <div className={styles.checkButton}>
            <PrimaryButton
              size="2"
              text={
                <FormattedMessage id="IPBlocklistForm.check-ip-address.button" />
              }
              onClick={onCheckIP}
              loading={checkingIP}
              disabled={checkingIP}
            />
          </div>
        </div>
      </FormField>
      {checkIPResult != null ? (
        checkIPResult.result ? (
          <Callout.Root color="red" variant="surface" size="1">
            <Callout.Icon>
              <ExclamationTriangleIcon />
            </Callout.Icon>
            <Callout.Text>
              <FormattedMessage
                id="IPBlocklistForm.check-ip-address.result.is-blocked"
                values={{
                  ipAddress: checkIPResult.ipAddress,
                  // eslint-disable-next-line react/no-unstable-nested-components
                  strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
                }}
              />
            </Callout.Text>
          </Callout.Root>
        ) : (
          <Callout.Root color="blue" variant="surface" size="1">
            <Callout.Icon>
              <InfoCircledIcon />
            </Callout.Icon>
            <Callout.Text>
              <FormattedMessage
                id="IPBlocklistForm.check-ip-address.result.is-not-blocked"
                values={{
                  ipAddress: checkIPResult.ipAddress,
                  // eslint-disable-next-line react/no-unstable-nested-components
                  strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
                }}
              />
            </Callout.Text>
          </Callout.Root>
        )
      ) : null}
    </div>
  );
}
