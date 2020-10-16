import React, { useMemo, useContext, useCallback } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react";

import ButtonWithLoading from "../../ButtonWithLoading";
import { destructiveTheme } from "../../theme";

export interface RemovePortalAdminInvitationConfirmationDialogData {
  invitationID: string;
  email: string;
}

export interface RemovePortalAdminInvitationConfirmationDialogProps {
  visible: boolean;
  data?: RemovePortalAdminInvitationConfirmationDialogData;
  deleteCollaboratorInvitation: (invitationID: string) => void;
  deletingCollaboratorInvitation: boolean;
  onDismiss: () => void;
}

const RemovePortalAdminInvitationConfirmationDialog: React.FC<RemovePortalAdminInvitationConfirmationDialogProps> = function RemovePortalAdminInvitationConfirmationDialog(
  props
) {
  const {
    visible,
    deleteCollaboratorInvitation,
    deletingCollaboratorInvitation,
    data,
    onDismiss: onDismissProps,
  } = props;

  const { renderToString } = useContext(Context);

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="RemovePortalAdminInvitationConfirmationDialog.title" />
      ),
      subText: renderToString(
        "RemovePortalAdminInvitationConfirmationDialog.message",
        { email: data?.email ?? "" }
      ),
    };
  }, [data?.email, renderToString]);

  const onConfirmClicked = useCallback(() => {
    deleteCollaboratorInvitation(data!.invitationID);
  }, [data, deleteCollaboratorInvitation]);

  const onDismiss = useCallback(() => {
    if (!deletingCollaboratorInvitation) {
      onDismissProps();
    }
  }, [onDismissProps, deletingCollaboratorInvitation]);

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={dialogContentProps}
      modalProps={{ isBlocking: deletingCollaboratorInvitation }}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <ButtonWithLoading
          onClick={onConfirmClicked}
          labelId="confirm"
          theme={destructiveTheme}
          loading={deletingCollaboratorInvitation}
          disabled={!visible}
        />
        <DefaultButton
          disabled={deletingCollaboratorInvitation || !visible}
          onClick={onDismiss}
        >
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default RemovePortalAdminInvitationConfirmationDialog;
