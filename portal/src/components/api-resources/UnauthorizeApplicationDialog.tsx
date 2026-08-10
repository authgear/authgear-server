import React, { useCallback, useContext } from "react";
import { Context, FormattedMessage } from "../../intl";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

export interface UnauthorizeApplicationDialogData {
  applicationName: string | null;
}

interface UnauthorizeApplicationDialogProps {
  data: UnauthorizeApplicationDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: UnauthorizeApplicationDialogData) => void;
}

export const UnauthorizeApplicationDialog: React.VFC<UnauthorizeApplicationDialogProps> =
  function UnauthorizeApplicationDialog(props) {
    const { onDismiss, onConfirm, data } = props;
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
        title={<FormattedMessage id="UnauthorizeApplicationDialog.title" />}
        description={
          <FormattedMessage
            id="UnauthorizeApplicationDialog.description"
            values={{
              applicationName: data?.applicationName ?? "Unknown Application",
            }}
          />
        }
        confirmText={renderToString("UnauthorizeApplicationDialog.unauthorize")}
        cancelText={renderToString("cancel")}
        onConfirm={handleConfirm}
        onCancel={onDismiss}
        confirmColor="red"
      />
    );
  };
