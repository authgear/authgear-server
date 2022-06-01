import React, { useCallback, useContext, useMemo } from "react";
import {
  DefaultButton,
  PrimaryButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useSetDisabledStatusMutation } from "./mutations/setDisabledStatusMutation";
import { useUnscheduleAccountDeletionMutation } from "./mutations/unscheduleAccountDeletion";
import ErrorDialog from "../../error/ErrorDialog";
import { extractRawID } from "../../util/graphql";

interface SetUserDisabledDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  userID: string;
  userDeleteAt: string | null;
  userIsDisabled: boolean;
  endUserAccountIdentifier: string | undefined;
}

const dialogStyles = { main: { minHeight: 0 } };

const SetUserDisabledDialog: React.FC<SetUserDisabledDialogProps> = React.memo(
  function SetUserDisabledDialog(props: SetUserDisabledDialogProps) {
    const {
      isHidden,
      onDismiss,
      userID,
      userDeleteAt,
      userIsDisabled,
      endUserAccountIdentifier,
    } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();
    const {
      setDisabledStatus,
      loading: setDisabledStatusLoading,
      error: setDisabledStatusError,
    } = useSetDisabledStatusMutation();
    const {
      unscheduleAccountDeletion,
      loading: unscheduleAccountDeletionLoading,
      error: unscheduleAccountDeletionError,
    } = useUnscheduleAccountDeletionMutation();

    const loading =
      setDisabledStatusLoading || unscheduleAccountDeletionLoading;
    const error = setDisabledStatusError || unscheduleAccountDeletionError;

    const onDialogDismiss = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      onDismiss();
    }, [loading, isHidden, onDismiss]);

    const onConfirm = useCallback(() => {
      if (loading || isHidden) {
        return;
      }
      if (userDeleteAt != null) {
        unscheduleAccountDeletion(userID).finally(() => onDismiss());
      } else {
        setDisabledStatus(userID, !userIsDisabled).finally(() => onDismiss());
      }
    }, [
      loading,
      isHidden,
      setDisabledStatus,
      unscheduleAccountDeletion,
      userID,
      userIsDisabled,
      userDeleteAt,
      onDismiss,
    ]);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      const args = {
        username: endUserAccountIdentifier ?? extractRawID(userID),
      };

      return userDeleteAt != null
        ? {
            title: renderToString("SetUserDisabledDialog.cancel-removal.title"),
            subText: renderToString(
              "SetUserDisabledDialog.cancel-removal.description",
              args
            ),
          }
        : userIsDisabled
        ? {
            title: renderToString("SetUserDisabledDialog.reenable-user.title"),
            subText: renderToString(
              "SetUserDisabledDialog.reenable-user.description",
              args
            ),
          }
        : {
            title: renderToString("SetUserDisabledDialog.disable-user.title"),
            subText: renderToString(
              "SetUserDisabledDialog.disable-user.description",
              args
            ),
          };
    }, [
      renderToString,
      userDeleteAt,
      userIsDisabled,
      endUserAccountIdentifier,
      userID,
    ]);

    const theme =
      userDeleteAt == null && !userIsDisabled
        ? themes.destructive
        : themes.main;

    const children =
      userDeleteAt != null ? (
        <FormattedMessage id="SetUserDisabledDialog.cancel-removal.label" />
      ) : userIsDisabled ? (
        <FormattedMessage id="reenable" />
      ) : (
        <FormattedMessage id="disable" />
      );

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
          styles={dialogStyles}
        >
          <DialogFooter>
            <PrimaryButton theme={theme} disabled={loading} onClick={onConfirm}>
              {children}
            </PrimaryButton>
            <DefaultButton onClick={onDialogDismiss} disabled={loading}>
              <FormattedMessage id="cancel" />
            </DefaultButton>
          </DialogFooter>
        </Dialog>
        <ErrorDialog error={error} />
      </>
    );
  }
);

export default SetUserDisabledDialog;
