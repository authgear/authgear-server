import React, { useCallback, useContext } from "react";
import { Context, FormattedMessage } from "../../intl";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

export interface DeleteScopeDialogData {
  scope: string;
  description: string | null;
}

interface DeleteScopeDialogProps {
  data: DeleteScopeDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: DeleteScopeDialogData) => void;
  isLoading: boolean;
}

export const DeleteScopeDialog: React.VFC<DeleteScopeDialogProps> =
  function DeleteScopeDialog(props) {
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
        title={<FormattedMessage id="DeleteScopeDialog.title" />}
        description={
          <FormattedMessage
            id="DeleteScopeDialog.description"
            values={{
              scope: data?.scope ?? "",
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
