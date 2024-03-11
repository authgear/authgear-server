import React, { useCallback, useContext } from "react";
import { Dialog, DialogFooter, IDialogContentProps } from "@fluentui/react";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ErrorDialog from "../../error/ErrorDialog";
import { useDeleteRoleMutation } from "./mutations/deleteRoleMutation";

interface DeleteRoleDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  roleID: string;
}

const DeleteRoleDialog: React.VFC<DeleteRoleDialogProps> =
  function DeleteRoleDialog(props) {
    const { isHidden, onDismiss, roleID } = props;
    const { renderToString } = useContext(Context);
    const dialogStyles = { main: { minHeight: 0 } };
    const { themes } = useSystemConfig();
    const { deleteRole, loading, error } = useDeleteRoleMutation();
    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss();
    }, [loading, isHidden, onDismiss]);

    const dialogContentProps: IDialogContentProps = {
      title: renderToString("DeleteRoleDialog.title"),
      subText: renderToString("DeleteRoleDialog.description"),
    };

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      deleteRole(roleID).finally(() => onDismiss());
    }, [loading, isHidden, deleteRole, roleID, onDismiss]);

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
          styles={dialogStyles}
        >
          <DialogFooter>
            <PrimaryButton
              theme={themes.destructive}
              disabled={loading}
              onClick={onConfirm}
              text={<FormattedMessage id="DeleteRoleDialog.button.confirm" />}
            />
            <DefaultButton
              onClick={onDialogDismiss}
              disabled={loading}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
        <ErrorDialog error={error} />
      </>
    );
  };

export default DeleteRoleDialog;
