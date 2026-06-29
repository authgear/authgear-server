import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";

import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";

export interface RemovePortalAdminConfirmationDialogData {
  userID: string;
  email: string;
}

export interface RemovePortalAdminConfirmationDialogProps {
  visible: boolean;
  data?: RemovePortalAdminConfirmationDialogData;
  deleteCollaborator: (userID: string) => void;
  deletingCollaborator: boolean;
  onDismiss: () => void;
}

const RemovePortalAdminConfirmationDialog: React.VFC<RemovePortalAdminConfirmationDialogProps> =
  function RemovePortalAdminConfirmationDialog(props) {
    const {
      visible,
      deleteCollaborator,
      deletingCollaborator,
      data,
      onDismiss: onDismissProps,
    } = props;

    const onConfirmClicked = useCallback(() => {
      if (data != null) {
        deleteCollaborator(data.userID);
      }
    }, [data, deleteCollaborator]);

    const onDismiss = useCallback(() => {
      if (!deletingCollaborator) {
        onDismissProps();
      }
    }, [onDismissProps, deletingCollaborator]);

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
          <FormattedMessage id="RemovePortalAdminConfirmationDialog.title" />
        }
        description={
          <FormattedMessage
            id="RemovePortalAdminConfirmationDialog.message"
            values={{ email: data?.email ?? "" }}
          />
        }
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onDismiss}
        loading={deletingCollaborator}
        confirmColor="red"
      />
    );
  };

export default RemovePortalAdminConfirmationDialog;
