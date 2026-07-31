import React, { useCallback, useContext } from "react";
import { Context, FormattedMessage } from "../../intl";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

export interface DeleteResourceDialogData {
  resourceURI: string;
  resourceName: string | null;
}

interface DeleteResourceDialogProps {
  data: DeleteResourceDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: DeleteResourceDialogData) => void;
  isLoading: boolean;
}

export const DeleteResourceDialog: React.VFC<DeleteResourceDialogProps> =
  function DeleteResourceDialog(props) {
    const { onDismiss, onConfirm, isLoading, data } = props;
    const { renderToString } = useContext(Context);

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss();
        }
      },
      [onDismiss]
    );

    const handleConfirm = useCallback(() => {
      if (data == null) {
        return;
      }
      onConfirm(data);
    }, [data, onConfirm]);

    return (
      <ConfirmationDialog
        open={data != null}
        onOpenChange={onOpenChange}
        title={<FormattedMessage id="DeleteResourceDialog.title" />}
        description={
          <FormattedMessage
            id="DeleteResourceDialog.description"
            values={{
              name: data?.resourceName ?? data?.resourceURI ?? "",
              // eslint-disable-next-line react/no-unstable-nested-components
              b: (chunks: React.ReactNode) => <b>{chunks}</b>,
            }}
          />
        }
        confirmText={renderToString("delete")}
        cancelText={renderToString("cancel")}
        onConfirm={handleConfirm}
        onCancel={onDismiss}
        loading={isLoading}
        confirmColor="red"
      />
    );
  };
