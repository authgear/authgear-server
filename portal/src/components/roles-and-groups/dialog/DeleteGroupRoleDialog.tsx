import React, { useCallback, useContext } from "react";
import { Context } from "../../../intl";
import { useGroupQuery } from "../../../graphql/adminapi/query/groupQuery";
import { useRemoveGroupFromRolesMutation } from "../../../graphql/adminapi/mutations/removeGroupFromRoles";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import { ConfirmationDialog } from "../../v2/ConfirmationDialog/ConfirmationDialog";
import ErrorDialog from "../../../error/ErrorDialog";

export interface DeleteGroupRoleDialogData {
  roleID: string;
  roleKey: string;
  roleName: string | null;
  groupID: string;
  groupName: string | null;
  groupKey: string;
}

interface DeleteGroupRoleDialogProps {
  data: DeleteGroupRoleDialogData | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
}

const DeleteGroupRoleDialog: React.VFC<DeleteGroupRoleDialogProps> =
  function DeleteGroupRoleDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);

    const { refetch: refetchGroup } = useGroupQuery(data?.groupID ?? "", {
      skip: true,
    });
    const { removeGroupFromRoles, loading, error } =
      useRemoveGroupFromRolesMutation();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteGroupRoleDialog.title");
    const description = renderToString("DeleteGroupRoleDialog.description", {
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
      removeGroupFromRoles(data.groupKey, [data.roleKey])
        .then(async () => {
          // Update the cache
          return refetchGroup({ groupID: data.groupID });
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
      refetchGroup,
      removeGroupFromRoles,
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

export default DeleteGroupRoleDialog;
