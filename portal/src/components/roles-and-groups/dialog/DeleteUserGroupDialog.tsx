import React, { useCallback, useContext } from "react";
import { Context } from "@oursky/react-messageformat";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import RolesAndGroupsBaseDeleteDialog from "./common/RolesAndGroupsBaseDeleteDialog";
import { useRemoveUserFromGroupsMutation } from "../../../graphql/adminapi/mutations/removeUserFromGroups";
import { useUserQuery } from "../../../graphql/adminapi/query/userQuery";
import { extractRawID } from "../../../util/graphql";

export interface DeleteUserGroupDialogData {
  userID: string;
  userFormattedName: string | null;
  userEndUserAccountID: string | null;
  groupID: string;
  groupName: string | null;
  groupKey: string;
}

interface DeleteUserGroupDialogProps {
  data: DeleteUserGroupDialogData | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
}

const DeleteUserGroupDialog: React.VFC<DeleteUserGroupDialogProps> =
  function DeleteUserGroupDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);

    const { refetch: refetchUser } = useUserQuery(data?.userID ?? "", {
      skip: true,
    });
    const { removeUserFromGroups, loading, error } =
      useRemoveUserFromGroupsMutation();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteUserGroupDialog.title");
    const subText = renderToString("DeleteUserGroupDialog.description", {
      groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
      userName:
        snapshot == null
          ? "Unknown"
          : snapshot.userFormattedName ??
            snapshot.userEndUserAccountID ??
            extractRawID(snapshot.userID),
    });
    const buttonText = renderToString("remove");

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      removeUserFromGroups(data.userID, [data.groupKey])
        .then(async () => {
          // Update the cache
          return refetchUser({ userID: data.userID });
        })
        .then(
          () => onDismiss(true),
          (e: unknown) => {
            onDismiss(false);
            throw e;
          }
        );
    }, [loading, isHidden, removeUserFromGroups, data, refetchUser, onDismiss]);

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

export default DeleteUserGroupDialog;
