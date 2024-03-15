import React, { useCallback, useContext } from "react";
import { Context } from "@oursky/react-messageformat";
import { useGroupQuery } from "../../../graphql/adminapi/query/groupQuery";
import { useRemoveGroupFromRolesMutation } from "../../../graphql/adminapi/mutations/removeGroupFromRoles";
import { useSnapshotData } from "../../../hook/useSnapshotData";
import RolesAndGroupsBaseDeleteDialog from "./common/RolesAndGroupsBaseDeleteDialog";

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

    const snapshot = useSnapshotData(data);
    const title = renderToString("DeleteGroupRoleDialog.title");
    const subText = renderToString("DeleteGroupRoleDialog.description", {
      groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
      roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
    });
    const buttonText = renderToString("remove");

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
          () => onDismiss(true),
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
    ]);

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

export default DeleteGroupRoleDialog;
