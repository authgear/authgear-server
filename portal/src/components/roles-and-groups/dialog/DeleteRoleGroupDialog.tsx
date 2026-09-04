import React, { useCallback, useContext } from "react";
import { Context } from "../../../intl";
import { useRemoveRoleFromGroupsMutation } from "../../../graphql/adminapi/mutations/removeRoleFromGroups";
import { useRoleQuery } from "../../../graphql/adminapi/query/roleQuery";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import { ConfirmationDialog } from "../../v2/ConfirmationDialog/ConfirmationDialog";
import ErrorDialog from "../../../error/ErrorDialog";

export interface DeleteRoleGroupDialogData {
  roleID: string;
  roleKey: string;
  roleName: string | null;
  groupID: string;
  groupName: string | null;
  groupKey: string;
}

interface DeleteRoleGroupDialogProps {
  data: DeleteRoleGroupDialogData | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
}

const DeleteRoleGroupDialog: React.VFC<DeleteRoleGroupDialogProps> =
  function DeleteRoleGroupDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);

    const { refetch: refetchRole } = useRoleQuery(data?.roleID ?? "", {
      skip: true,
    });
    const { removeRoleFromGroups, loading, error } =
      useRemoveRoleFromGroupsMutation();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteRoleGroupDialog.title");
    const description = renderToString("DeleteRoleGroupDialog.description", {
      groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
      roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
    });
    const confirmText = renderToString("remove");

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
      removeRoleFromGroups(data.roleKey, [data.groupKey])
        .then(async () => {
          // Update the cache
          return refetchRole({ roleID: data.roleID });
        })
        .then(
          () => {
            onDismiss(true);
            onDismissed?.();
          },
          (e: unknown) => {
            onDismiss(false);
            throw e;
          }
        );
    }, [
      loading,
      isHidden,
      refetchRole,
      removeRoleFromGroups,
      data,
      onDismiss,
      onDismissed,
    ]);

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

export default DeleteRoleGroupDialog;
