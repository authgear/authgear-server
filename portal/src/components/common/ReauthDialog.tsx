import { Dialog, DialogFooter, IDialogProps } from "@fluentui/react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "../../intl";
import React, { useContext, useMemo } from "react";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

export interface ReauthDialogProps {
  isHidden: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ReauthDialog({
  isHidden,
  onCancel,
  onConfirm,
}: ReauthDialogProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  return (
    <Dialog
      hidden={isHidden}
      dialogContentProps={useMemo<IDialogProps["dialogContentProps"]>(() => {
        return {
          title: <FormattedMessage id="ReauthDialog.title" />,
          subText: renderToString("ReauthDialog.description"),
        };
      }, [renderToString])}
      onDismiss={onCancel}
    >
      <DialogFooter>
        <PrimaryButton
          onClick={onConfirm}
          text={<FormattedMessage id="confirm" />}
        />
        <DefaultButton
          onClick={onCancel}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
}
