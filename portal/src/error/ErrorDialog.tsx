import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, DialogFooter, PrimaryButton } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import { GenericErrorHandlingRule, useGenericError } from "./useGenericError";

interface ErrorDialogProps {
  error: unknown;
  rules: GenericErrorHandlingRule[];
  fallbackErrorMessageID?: string;
}

const ErrorDialog: React.FC<ErrorDialogProps> = function ErrorDialog(
  props: ErrorDialogProps
) {
  const { error, rules, fallbackErrorMessageID } = props;
  const errorMessage = useGenericError(error, rules, fallbackErrorMessageID);

  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (errorMessage != null) {
      setVisible(true);
    }
  }, [errorMessage]);

  const errorDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="error" />,
      subText: errorMessage,
    };
  }, [errorMessage]);

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
        <PrimaryButton onClick={onDismiss}>
          <FormattedMessage id="ok" />
        </PrimaryButton>
      </DialogFooter>
    </Dialog>
  );
};

export default ErrorDialog;
