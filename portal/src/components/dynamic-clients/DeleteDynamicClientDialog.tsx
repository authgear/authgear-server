import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";

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
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const dialogContentProps = useMemo(
      () => ({
        title: renderToString("DeleteDynamicClientDialog.title"),
        subText: renderToString("DeleteDynamicClientDialog.description", {
          clientName: data?.clientName ?? "",
        }),
      }),
      [renderToString, data?.clientName]
    );

    const onConfirmClicked = useCallback(() => {
      if (data != null) {
        onConfirm(data);
      }
    }, [data, onConfirm]);

    return (
      <Dialog
        hidden={data == null}
        dialogContentProps={dialogContentProps}
        modalProps={{ isBlocking: isLoading }}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <ButtonWithLoading
            theme={themes.destructive}
            loading={isLoading}
            onClick={onConfirmClicked}
            disabled={data == null}
            labelId="DeleteDynamicClientDialog.confirm"
          />
          <DefaultButton
            onClick={onDismiss}
            disabled={isLoading || data == null}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };
