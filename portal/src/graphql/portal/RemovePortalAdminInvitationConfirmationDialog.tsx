import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";

import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";

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

const RemovePortalAdminInvitationConfirmationDialog: React.VFC<RemovePortalAdminInvitationConfirmationDialogProps> =
  function RemovePortalAdminInvitationConfirmationDialog(props) {
    const {
      visible,
      deleteCollaboratorInvitation,
      deletingCollaboratorInvitation,
      data,
      onDismiss: onDismissProps,
    } = props;

    const onConfirmClicked = useCallback(() => {
      if (data != null) {
        deleteCollaboratorInvitation(data.invitationID);
      }
    }, [data, deleteCollaboratorInvitation]);

    const onDismiss = useCallback(() => {
      if (!deletingCollaboratorInvitation) {
        onDismissProps();
      }
    }, [onDismissProps, deletingCollaboratorInvitation]);

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss();
        }
      },
      [onDismiss]
    );

    return (
      <ConfirmationDialog
        open={visible}
        onOpenChange={onOpenChange}
        title={
          <FormattedMessage id="RemovePortalAdminInvitationConfirmationDialog.title" />
        }
        description={
          <FormattedMessage
            id="RemovePortalAdminInvitationConfirmationDialog.message"
            values={{ email: data?.email ?? "" }}
          />
        }
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onDismiss}
        loading={deletingCollaboratorInvitation}
        confirmColor="red"
      />
    );
  };

export default RemovePortalAdminInvitationConfirmationDialog;
