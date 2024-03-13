import React, { useCallback, useContext, useMemo } from "react";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
} from "@fluentui/react";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ErrorDialog from "../../error/ErrorDialog";
import { useRemoveRoleFromGroupsMutation } from "../../graphql/adminapi/mutations/removeRoleFromGroups";
import { useRoleQuery } from "../../graphql/adminapi/query/roleQuery";
import { useSnapshotData } from "../../hook/useSnapshotData";

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

const dialogStyles = { main: { minHeight: 0 } };

const DeleteRoleGroupDialog: React.VFC<DeleteRoleGroupDialogProps> =
  function DeleteRoleGroupDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const { refetch: refetchRole } = useRoleQuery(data?.roleID ?? "", {
      skip: true,
    });
    const { removeRoleFromGroups, loading, error } =
      useRemoveRoleFromGroupsMutation();
    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);
    const dialogContentProps: IDialogContentProps = {
      title: renderToString("DeleteRoleGroupDialog.title"),
      subText: renderToString("DeleteRoleGroupDialog.description", {
        groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
        roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
      }),
    };

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
          () => onDismiss(true),
          (e: unknown) => {
            onDismiss(false);
            throw e;
          }
        );
    }, [loading, isHidden, refetchRole, removeRoleFromGroups, data, onDismiss]);

    const modalProps = useMemo((): IModalProps => {
      return {
        onDismissed,
      };
    }, [onDismissed]);

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          modalProps={modalProps}
          dialogContentProps={dialogContentProps}
          styles={dialogStyles}
        >
          <DialogFooter>
            <PrimaryButton
              theme={themes.destructive}
              disabled={loading}
              onClick={onConfirm}
              text={<FormattedMessage id="remove" />}
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

export default DeleteRoleGroupDialog;
