import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
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
import { useDeleteRoleMutation } from "./mutations/deleteRoleMutation";

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

function useSnapshotData(data: DeleteRoleDialogData | null) {
  // Keep the latest non-null data, because the dialog has transition animation before dismiss.
  // During the transition, we still need the data. However, the parent may already changed the props.
  const [snapshot, setSnapshot] = useState<DeleteRoleDialogData | null>(data);
  useEffect(() => {
    if (data !== null) {
      setSnapshot(data);
    }
  }, [data]);
  return snapshot;
}

const DeleteRoleDialog: React.VFC<DeleteRoleDialogProps> =
  function DeleteRoleDialog(props) {
    const { onDismiss, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
    const dialogStyles = { main: { minHeight: 0 } };
    const { themes } = useSystemConfig();
    const { deleteRole, loading, error } = useDeleteRoleMutation();
    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    const snapshot = useSnapshotData(data);
    const dialogContentProps: IDialogContentProps = {
      title: renderToString("DeleteRoleDialog.title"),
      subText: renderToString("DeleteRoleDialog.description", {
        roleName: snapshot?.roleName ?? snapshot?.roleKey ?? "Unknown",
      }),
    };

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
              text={<FormattedMessage id="DeleteRoleDialog.button.confirm" />}
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

export default DeleteRoleDialog;
