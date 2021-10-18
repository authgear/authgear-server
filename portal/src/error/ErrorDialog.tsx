import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { Dialog, DialogFooter, PrimaryButton } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import {
  ErrorParseRule,
  parseAPIErrors,
  parseRawError,
  renderErrors,
} from "./parse";

interface ErrorDialogProps {
  error: unknown;
  rules?: ErrorParseRule[];
  fallbackErrorMessageID?: string;
}

const ErrorDialog: React.FC<ErrorDialogProps> = function ErrorDialog(
  props: ErrorDialogProps
) {
  const { error, rules, fallbackErrorMessageID } = props;
  const { renderToString } = useContext(Context);

  const { topErrors } = useMemo(() => {
    const apiErrors = parseRawError(error);
    return parseAPIErrors(apiErrors, [], rules ?? [], fallbackErrorMessageID);
  }, [error, rules, fallbackErrorMessageID]);

  const message = useMemo(
    () => renderErrors(topErrors, renderToString),
    [topErrors, renderToString]
  );

  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (message != null) {
      setVisible(true);
    }
  }, [message]);

  const errorDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="error" />,
      subText: message,
    };
  }, [message]);

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
