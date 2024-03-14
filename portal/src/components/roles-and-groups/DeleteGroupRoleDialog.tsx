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
import { useGroupQuery } from "../../graphql/adminapi/query/groupQuery";
import { useRemoveGroupFromRolesMutation } from "../../graphql/adminapi/mutations/removeGroupFromRoles";
import { useSnapshotData } from "../../hook/useSnapshotData";

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

const dialogStyles = { main: { minHeight: 0 } };

const DeleteGroupRoleDialog: React.VFC<DeleteGroupRoleDialogProps> =
  function DeleteGroupRoleDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const { refetch: refetchGroup } = useGroupQuery(data?.groupID ?? "", {
      skip: true,
    });
    const { removeGroupFromRoles, loading, error } =
      useRemoveGroupFromRolesMutation();
    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    const snapshot = useSnapshotData(data);
    const dialogContentProps: IDialogContentProps = {
      title: renderToString("DeleteGroupRoleDialog.title"),
      subText: renderToString("DeleteGroupRoleDialog.description", {
        groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
        roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
      }),
    };

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

export default DeleteGroupRoleDialog;
