import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Dialog } from "@radix-ui/themes";
import { FormattedMessage, Values } from "../intl";

import { ErrorParseRule, parseAPIErrors, parseRawError } from "./parse";
import { PrimaryButton } from "../components/v2/Button/PrimaryButton/PrimaryButton";
import ErrorRenderer from "../ErrorRenderer";
import styles from "./ErrorDialog.module.css";

interface ErrorDialogProps {
  titleMessageID?: string;
  error: unknown;
  rules?: ErrorParseRule[];
  fallbackErrorMessageID?: string;
  fallbackErrorMessageValues?: Values;
}

const ErrorDialog: React.VFC<ErrorDialogProps> = function ErrorDialog(
  props: ErrorDialogProps
) {
  const {
    titleMessageID,
    error,
    rules,
    fallbackErrorMessageID,
    fallbackErrorMessageValues,
  } = props;

  const { topErrors } = useMemo(() => {
    const apiErrors = parseRawError(error);
    return parseAPIErrors(
      apiErrors,
      [],
      rules ?? [],
      fallbackErrorMessageID,
      fallbackErrorMessageValues
    );
  }, [error, rules, fallbackErrorMessageID, fallbackErrorMessageValues]);

  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (error != null) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setVisible(true);
    }
  }, [error]);

  const onDismiss = useCallback(() => {
    setVisible(false);
  }, []);

  const onOpenChange = useCallback((open: boolean) => {
    if (!open) {
      setVisible(false);
    }
  }, []);

  return (
    <Dialog.Root open={visible} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="400px" size="3">
        <Dialog.Title>
          <FormattedMessage id={titleMessageID ?? "error"} />
        </Dialog.Title>
        <Dialog.Description size="2">
          <ErrorRenderer errors={topErrors} />
        </Dialog.Description>
        <div className={styles.actions}>
          <PrimaryButton
            size="2"
            onClick={onDismiss}
            text={<FormattedMessage id="ok" />}
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
};

export default ErrorDialog;
