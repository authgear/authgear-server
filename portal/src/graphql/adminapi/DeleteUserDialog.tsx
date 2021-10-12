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
import { useDeleteUserMutation } from "./mutations/deleteUserMutation";

interface DeleteUserDialogProps {
  isHidden: boolean;
  onDismiss: (deletedUser: boolean) => void;
  userID: string;
  endUserAccountIdentifier: string | undefined;
}

const DeleteUserDialog: React.FC<DeleteUserDialogProps> = React.memo(
  function DeleteUserDialog(props: DeleteUserDialogProps) {
    const { isHidden, onDismiss, userID, endUserAccountIdentifier } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
    const { deleteUser, loading, error } = useDeleteUserMutation();

    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      deleteUser(userID)
        .then(() => onDismiss(true))
        .catch(() => onDismiss(false));
    }, [loading, isHidden, deleteUser, userID, onDismiss]);

    const dialogContentProps: IDialogContentProps = useMemo(
      () => ({
        title: renderToString("DeleteUserDialog.title"),
        subText: renderToString("DeleteUserDialog.text", {
          username: endUserAccountIdentifier ?? userID,
        }),
      }),
      [renderToString, endUserAccountIdentifier, userID]
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
          error={error}
          fallbackErrorMessageID="DeleteUserDialog.generic-error"
        />
      </>
    );
  }
);

export default DeleteUserDialog;
