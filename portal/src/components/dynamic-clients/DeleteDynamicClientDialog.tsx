import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

export interface DeleteDynamicClientDialogData {
  clientID: string;
  clientName: string;
}

export interface DeleteDynamicClientDialogProps {
  data: DeleteDynamicClientDialogData | null;
  isLoading: boolean;
  onConfirm: (data: DeleteDynamicClientDialogData) => void;
  onDismiss: () => void;
}

export const DeleteDynamicClientDialog: React.VFC<DeleteDynamicClientDialogProps> =
  function DeleteDynamicClientDialog({
    data,
    isLoading,
    onConfirm,
    onDismiss,
  }) {
    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open && !isLoading) {
          onDismiss();
        }
      },
      [isLoading, onDismiss]
    );

    const onCancel = useCallback(() => {
      if (!isLoading) {
        onDismiss();
      }
    }, [isLoading, onDismiss]);

    const onConfirmClicked = useCallback(() => {
      if (data != null) {
        onConfirm(data);
      }
    }, [data, onConfirm]);

    return (
      <ConfirmationDialog
        open={data != null}
        onOpenChange={onOpenChange}
        title={<FormattedMessage id="DeleteDynamicClientDialog.title" />}
        description={
          <FormattedMessage
            id="DeleteDynamicClientDialog.description"
            values={{ clientName: data?.clientName ?? "" }}
          />
        }
        confirmText={
          <FormattedMessage id="DeleteDynamicClientDialog.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onCancel}
        loading={isLoading}
        confirmColor="red"
      />
    );
  };
