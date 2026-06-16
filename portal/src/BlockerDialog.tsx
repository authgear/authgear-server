import React from "react";
import { FormattedMessage } from "./intl";
import { ConfirmationDialog } from "./components/v2/ConfirmationDialog/ConfirmationDialog";

export interface BlockerDialogProps {
  open: boolean;
  contentTitleId: string;
  contentSubTextId: string;
  contentConfirmId?: string;
  contentCancelId?: string;
  onDialogConfirm?: () => void;
  onDialogDismiss?: () => void;
}

const BlockerDialog: React.VFC<BlockerDialogProps> = function BlockerDialog(
  props
) {
  const {
    open,
    contentTitleId,
    contentSubTextId,
    contentConfirmId,
    contentCancelId,
    onDialogConfirm,
    onDialogDismiss,
  } = props;

  return (
    <ConfirmationDialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onDialogDismiss?.();
        }
      }}
      title={<FormattedMessage id={contentTitleId} />}
      description={<FormattedMessage id={contentSubTextId} />}
      confirmText={<FormattedMessage id={contentConfirmId ?? "confirm"} />}
      cancelText={<FormattedMessage id={contentCancelId ?? "cancel"} />}
      confirmColor="red"
      onConfirm={() => onDialogConfirm?.()}
      onCancel={() => onDialogDismiss?.()}
    />
  );
};

export default BlockerDialog;
