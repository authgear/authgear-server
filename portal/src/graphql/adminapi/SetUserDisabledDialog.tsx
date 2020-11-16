import React, { useCallback, useContext, useMemo } from "react";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
  PrimaryButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

interface SetUserDisabledDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  isDisablingUser: boolean;
  userID: string;
  username: string | null;
}

const dialogStyles = { main: { minHeight: 0 } };

const SetUserDisabledDialog: React.FC<SetUserDisabledDialogProps> = function SetUserDisabledDialog(
  props: SetUserDisabledDialogProps
) {
  const { isHidden, onDismiss, isDisablingUser, userID, username } = props;
  const { renderToString } = useContext(Context);

  const setUserDisabled = useCallback(
    (userID?: string) => {
      if (userID == null) {
        onDismiss();
        return;
      }
      // TODO: call mutation
      onDismiss();
    },
    [onDismiss]
  );

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
    <Dialog
      hidden={isHidden}
      onDismiss={onDismiss}
      dialogContentProps={dialogContentProps}
      styles={dialogStyles}
    >
      <DialogFooter>
        <PrimaryButton onClick={() => setUserDisabled(userID)}>
          <FormattedMessage id="confirm" />
        </PrimaryButton>

        <DefaultButton onClick={onDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default SetUserDisabledDialog;
