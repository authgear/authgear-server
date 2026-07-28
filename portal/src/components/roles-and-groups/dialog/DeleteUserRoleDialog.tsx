import React, { useCallback, useContext } from "react";
import { Context } from "../../../intl";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import { ConfirmationDialog } from "../../v2/ConfirmationDialog/ConfirmationDialog";
import ErrorDialog from "../../../error/ErrorDialog";
import { useRemoveUserFromRolesMutation } from "../../../graphql/adminapi/mutations/removeUserFromRoles";
import { useUserQuery } from "../../../graphql/adminapi/query/userQuery";
import { extractRawID } from "../../../util/graphql";

export interface DeleteUserRoleDialogData {
  userID: string;
  userFormattedName: string | null;
  userEndUserAccountID: string | null;
  roleID: string;
  roleName: string | null;
  roleKey: string;
}

interface DeleteUserRoleDialogProps {
  data: DeleteUserRoleDialogData | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
}

const DeleteUserRoleDialog: React.VFC<DeleteUserRoleDialogProps> =
  function DeleteUserRoleDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);

    const { refetch: refetchUser } = useUserQuery(data?.userID ?? "", {
      skip: true,
    });
    const { removeUserFromRoles, loading, error } =
      useRemoveUserFromRolesMutation();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteUserRoleDialog.title");
    const description = renderToString("DeleteUserRoleDialog.description", {
      roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
      userName:
        snapshot == null
          ? "Unknown"
          : snapshot.userFormattedName ??
            snapshot.userEndUserAccountID ??
            extractRawID(snapshot.userID),
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
      removeUserFromRoles(data.userID, [data.roleKey])
        .then(async () => {
          // Update the cache
          return refetchUser({ userID: data.userID });
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
      removeUserFromRoles,
      data,
      refetchUser,
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

export default DeleteUserRoleDialog;
