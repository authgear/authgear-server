import React, { useCallback, useContext, useMemo } from "react";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";

interface DeleteUserDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  userID: string;
  username: string | null;
}

const DeleteUserDialog: React.FC<DeleteUserDialogProps> = React.memo(
  function DeleteUserDialog(props: DeleteUserDialogProps) {
    const { isHidden, onDismiss, userID, username } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const onDialogDismiss = useCallback(() => {
      if (isHidden) {
        return;
      }
      onDismiss();
    }, [isHidden, onDismiss]);

    const onConfirm = useCallback(() => {}, []);

    const dialogContentProps: IDialogContentProps = useMemo(
      () => ({
        title: renderToString("DeleteUserDialog.title"),
        subText: renderToString("DeleteUserDialog.text", {
          username: username ?? userID,
        }),
      }),
      [renderToString, username, userID]
    );

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
        >
          <DialogFooter>
            <ButtonWithLoading
              theme={themes.destructive}
              onClick={onConfirm}
              labelId={"DeleteUserDialog.action"}
              loading={false}
            />

            <DefaultButton onClick={onDialogDismiss} disabled={false}>
              <FormattedMessage id="cancel" />
            </DefaultButton>
          </DialogFooter>
        </Dialog>
        <ErrorDialog
          rules={[]}
          error={null}
          fallbackErrorMessageID="DeleteUserDialog.generic-error"
        />
      </>
    );
  }
);

export default DeleteUserDialog;
