import React, { useCallback, useContext } from "react";
import { Context } from "../../../intl";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import { useDeleteGroupMutation } from "../../../graphql/adminapi/mutations/deleteGroupMutation";
import RolesAndGroupsBaseDeleteDialog from "./common/RolesAndGroupsBaseDeleteDialog";

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
    const subText = renderToString("DeleteGroupDialog.description", {
      groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
    });
    const buttonText = renderToString("DeleteGroupDialog.button.confirm");

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      deleteGroup(data.groupID).then(
        () => onDismiss(true),
        (e: unknown) => {
          onDismiss(false);
          throw e;
        }
      );
    }, [loading, isHidden, deleteGroup, data, onDismiss]);

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

export default DeleteGroupDialog;
