import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter, IDialogContentProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ErrorDialog from "../../error/ErrorDialog";
import { useDeleteUserMutation } from "./mutations/deleteUserMutation";
import { useScheduleAccountDeletionMutation } from "./mutations/scheduleAccountDeletion";
import { extractRawID } from "../../util/graphql";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

interface DeleteUserDialogProps {
  isHidden: boolean;
  onDismiss: (deletedUser: boolean) => void;
  userID: string;
  userDeleteAt: string | null;
  endUserAccountIdentifier: string | undefined;
}

const DeleteUserDialog: React.VFC<DeleteUserDialogProps> = React.memo(
  function DeleteUserDialog(props: DeleteUserDialogProps) {
    const {
      isHidden,
      onDismiss,
      userID,
      userDeleteAt,
      endUserAccountIdentifier,
    } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
    const {
      deleteUser,
      loading: deleteUserLoading,
      error: deleteUserError,
    } = useDeleteUserMutation();
    const {
      scheduleAccountDeletion,
      loading: scheduleAccountDeletionLoading,
      error: scheduleAccountDeletionError,
    } = useScheduleAccountDeletionMutation();

    const loading = deleteUserLoading || scheduleAccountDeletionLoading;
    const error = deleteUserError || scheduleAccountDeletionError;

    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss(false);
    }, [loading, isHidden, onDismiss]);

    const onClickRemove = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      deleteUser(userID)
        .then(() => onDismiss(true))
        .catch(() => onDismiss(false));
    }, [loading, isHidden, deleteUser, userID, onDismiss]);

    const onClickScheduleDeletion = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      scheduleAccountDeletion(userID)
        .then(() => onDismiss(false))
        .catch(() => onDismiss(false));
      onDismiss(false);
    }, [loading, isHidden, scheduleAccountDeletion, userID, onDismiss]);

    const dialogContentProps: IDialogContentProps = useMemo(
      () => ({
        title: renderToString("DeleteUserDialog.title"),
        subText: renderToString("DeleteUserDialog.text", {
          username: endUserAccountIdentifier ?? extractRawID(userID),
        }),
      }),
      [renderToString, endUserAccountIdentifier, userID]
    );

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
          minWidth={500}
        >
          <DialogFooter>
            {userDeleteAt == null ? (
              <PrimaryButton
                onClick={onClickScheduleDeletion}
                disabled={loading}
                text={
                  <FormattedMessage id="DeleteUserDialog.label.schedule-removal" />
                }
              />
            ) : null}
            <PrimaryButton
              theme={themes.destructive}
              onClick={onClickRemove}
              disabled={loading}
              text={
                <FormattedMessage id="DeleteUserDialog.label.remove-immediately" />
              }
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
  }
);

export default DeleteUserDialog;
