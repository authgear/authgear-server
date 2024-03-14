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
import { useSnapshotData } from "../../hook/useSnapshotData";
import { useDeleteGroupMutation } from "../../graphql/adminapi/mutations/deleteGroupMutation";

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
    const dialogStyles = { main: { minHeight: 0 } };
    const { themes } = useSystemConfig();
    const { deleteGroup, loading, error } = useDeleteGroupMutation();
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
      title: renderToString("DeleteGroupDialog.title"),
      subText: renderToString("DeleteGroupDialog.description", {
        groupName: snapshot?.groupName ?? snapshot?.groupKey ?? "Unknown",
      }),
    };

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
              text={<FormattedMessage id="DeleteGroupDialog.button.confirm" />}
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

export default DeleteGroupDialog;
