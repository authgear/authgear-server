import React, { useCallback, useContext } from "react";
import { Context } from "@oursky/react-messageformat";
import { useDeleteRoleMutation } from "../../../graphql/adminapi/mutations/deleteRoleMutation";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import RolesAndGroupsBaseDeleteDialog from "./common/RolesAndGroupsBaseDeleteDialog";

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
    const subText = renderToString("DeleteRoleDialog.description", {
      roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
    });
    const buttonText = renderToString("DeleteRoleDialog.button.confirm");

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      deleteRole(data.roleID).then(
        () => onDismiss(true),
        (e: unknown) => {
          onDismiss(false);
          throw e;
        }
      );
    }, [loading, isHidden, deleteRole, data, onDismiss]);

    return (
      <RolesAndGroupsBaseDeleteDialog
        data={snapshot}
        loading={loading}
        error={error}
        title={title}
        // eslint-disable-next-line react/forbid-component-props
        subText={subText}
        buttonText={buttonText}
        isHidden={isHidden}
        onDismiss={onDismiss}
        onDismissed={onDismissed}
        onConfirm={onConfirm}
      />
    );
  };

export default DeleteRoleDialog;
