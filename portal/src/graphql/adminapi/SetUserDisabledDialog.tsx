import React, { useCallback, useContext, useMemo } from "react";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSetDisabledStatusMutation } from "./mutations/setDisabledStatusMutation";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";

interface SetUserDisabledDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  isDisablingUser: boolean;
  userID: string;
  username: string | null;
}

const dialogStyles = { main: { minHeight: 0 } };

const SetUserDisabledDialog: React.FC<SetUserDisabledDialogProps> = React.memo(
  function SetUserDisabledDialog(props: SetUserDisabledDialogProps) {
    const { isHidden, onDismiss, isDisablingUser, userID, username } = props;
    const { renderToString } = useContext(Context);
    const { setDisabledStatus, loading, error } = useSetDisabledStatusMutation(
      userID
    );

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
      setDisabledStatus(isDisablingUser).finally(() => onDismiss());
    }, [loading, isHidden, setDisabledStatus, isDisablingUser, onDismiss]);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      return isDisablingUser
        ? {
            title: renderToString("SetUserDisabledDialog.disableUser.title"),
            subText: renderToString("SetUserDisabledDialog.disableUser.text", {
              username: username ?? userID,
            }),
          }
        : {
            title: renderToString("SetUserDisabledDialog.enableUser.title"),
            subText: renderToString("SetUserDisabledDialog.enableUser.text", {
              username: username ?? userID,
            }),
          };
    }, [renderToString, isDisablingUser, username, userID]);

    return (
      <>
        <Dialog
          hidden={isHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
          styles={dialogStyles}
        >
          <DialogFooter>
            <ButtonWithLoading
              onClick={onConfirm}
              labelId="confirm"
              loading={loading}
            />

            <DefaultButton onClick={onDialogDismiss} disabled={loading}>
              <FormattedMessage id="cancel" />
            </DefaultButton>
          </DialogFooter>
        </Dialog>
        <ErrorDialog
          rules={[]}
          error={error}
          fallbackErrorMessageID="SetUserDisabledDialog.generic-error"
        />
      </>
    );
  }
);

export default SetUserDisabledDialog;
