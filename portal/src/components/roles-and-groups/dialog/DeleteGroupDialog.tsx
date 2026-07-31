import React, { useCallback, useContext } from "react";
import { Context } from "../../../intl";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import { useDeleteGroupMutation } from "../../../graphql/adminapi/mutations/deleteGroupMutation";
import { ConfirmationDialog } from "../../v2/ConfirmationDialog/ConfirmationDialog";
import ErrorDialog from "../../../error/ErrorDialog";

export interface DeleteGroupDialogData {
  groupID: string;
  groupName: string | null;
  groupKey: string;
}

interface DeleteGroupDialogProps {
  data: DeleteGroupDialogData | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
}

const DeleteGroupDialog: React.VFC<DeleteGroupDialogProps> =
  function DeleteGroupDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
    const { deleteGroup, loading, error } = useDeleteGroupMutation();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteGroupDialog.title");
    const description = renderToString("DeleteGroupDialog.description", {
      groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
    });
    const confirmText = renderToString("DeleteGroupDialog.button.confirm");

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
      deleteGroup(data.groupID).then(
        () => {
          onDismiss(true);
          onDismissed?.();
        },
        (e: unknown) => {
          onDismiss(false);
          throw e;
        }
      );
    }, [loading, isHidden, deleteGroup, data, onDismiss, onDismissed]);

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

export default DeleteGroupDialog;
