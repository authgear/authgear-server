import React, { useCallback, useContext } from "react";
import { Context } from "../../../intl";
import { useDeleteRoleMutation } from "../../../graphql/adminapi/mutations/deleteRoleMutation";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import { ConfirmationDialog } from "../../v2/ConfirmationDialog/ConfirmationDialog";
import ErrorDialog from "../../../error/ErrorDialog";

export interface DeleteRoleDialogData {
  roleID: string;
  roleName: string | null;
  roleKey: string;
}

interface DeleteRoleDialogProps {
  data: DeleteRoleDialogData | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
}

const DeleteRoleDialog: React.VFC<DeleteRoleDialogProps> =
  function DeleteRoleDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
    const { deleteRole, loading, error } = useDeleteRoleMutation();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteRoleDialog.title");
    const description = renderToString("DeleteRoleDialog.description", {
      roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
    });
    const confirmText = renderToString("DeleteRoleDialog.button.confirm");

    const onCancel = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onCancel();
        }
      },
      [onCancel]
    );

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      deleteRole(data.roleID).then(
        () => {
          onDismiss(true);
          onDismissed?.();
        },
        (e: unknown) => {
          onDismiss(false);
          throw e;
        }
      );
    }, [loading, isHidden, deleteRole, data, onDismiss, onDismissed]);

    return (
      <>
        <ConfirmationDialog
          open={!isHidden}
          onOpenChange={onOpenChange}
          title={title}
          description={description}
          confirmText={confirmText}
          cancelText={renderToString("cancel")}
          onConfirm={onConfirm}
          onCancel={onCancel}
          loading={loading}
          confirmColor="red"
        />
        <ErrorDialog error={error} />
      </>
    );
  };

export default DeleteRoleDialog;
