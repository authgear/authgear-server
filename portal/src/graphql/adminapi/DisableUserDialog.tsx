import React, { useCallback, useContext, useMemo } from "react";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
  PrimaryButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

interface DisableUserDialogProps {
  onDismiss: () => void;
  userID: string;
  username: string | null;
}

const dialogStyles = { main: { minWidth: "400px !important", minHeight: 0 } };

const DisableUserDialog: React.FC<DisableUserDialogProps> = function DisableUserDialog(
  props: DisableUserDialogProps
) {
  const { onDismiss, userID, username } = props;
  const { renderToString } = useContext(Context);

  const disableUser = useCallback(
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
    return {
      title: renderToString("DisableUserDialog.title"),
      subText: renderToString("DisableUserDialog.text", {
        username: username ?? userID,
      }),
    };
  }, [renderToString, username, userID]);

  return (
    <Dialog
      hidden={false}
      onDismiss={onDismiss}
      dialogContentProps={dialogContentProps}
      styles={dialogStyles}
    >
      <DialogFooter>
        <PrimaryButton onClick={() => disableUser(userID)}>
          <FormattedMessage id="confirm" />
        </PrimaryButton>

        <DefaultButton onClick={onDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default DisableUserDialog;
