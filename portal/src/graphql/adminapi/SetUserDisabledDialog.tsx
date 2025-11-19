import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter, IDialogContentProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useSetDisabledStatusMutation } from "./mutations/setDisabledStatusMutation";
import ErrorDialog from "../../error/ErrorDialog";
import { extractRawID } from "../../util/graphql";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

interface SetUserDisabledDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  userID: string;
  userIsDisabled: boolean;
  endUserAccountIdentifier: string | undefined;
}

const dialogStyles = { main: { minHeight: 0 } };

const SetUserDisabledDialog: React.VFC<SetUserDisabledDialogProps> = React.memo(
  function SetUserDisabledDialog(props: SetUserDisabledDialogProps) {
    const {
      isHidden,
      onDismiss,
      userID,
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

    const loading = setDisabledStatusLoading;
    const error = setDisabledStatusError;

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
      setDisabledStatus(userID, !userIsDisabled).finally(() => onDismiss());
    }, [
      loading,
      isHidden,
      setDisabledStatus,
      userID,
      userIsDisabled,
      onDismiss,
    ]);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      const args = {
        username: endUserAccountIdentifier ?? extractRawID(userID),
      };

      return userIsDisabled
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
    }, [renderToString, userIsDisabled, endUserAccountIdentifier, userID]);

    const theme = userIsDisabled ? themes.main : themes.destructive;
    const children = userIsDisabled ? (
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
            <PrimaryButton
              theme={theme}
              disabled={loading}
              onClick={onConfirm}
              text={children}
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

export default SetUserDisabledDialog;
