import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, DialogFooter, IDialogContentProps } from "@fluentui/react";
import { FormattedMessage, Values } from "../intl";

import { ErrorParseRule, parseAPIErrors, parseRawError } from "./parse";
import PrimaryButton from "../PrimaryButton";
import ErrorRenderer from "../ErrorRenderer";

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
      setVisible(true);
    }
  }, [error]);

  // @ts-expect-error
  const errorDialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id={titleMessageID ?? "error"} />,
      subText: <ErrorRenderer errors={topErrors} />,
    };
  }, [titleMessageID, topErrors]);

  const onDismiss = useCallback(() => {
    setVisible(false);
  }, []);

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={errorDialogContentProps}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <PrimaryButton
          onClick={onDismiss}
          text={<FormattedMessage id="ok" />}
        />
      </DialogFooter>
    </Dialog>
  );
};

export default ErrorDialog;
