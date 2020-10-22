import React, { useMemo, useContext, useCallback } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react";

import ButtonWithLoading from "../../ButtonWithLoading";
import { useSystemConfig } from "../../context/SystemConfigContext";

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

const RemovePortalAdminConfirmationDialog: React.FC<RemovePortalAdminConfirmationDialogProps> = function RemovePortalAdminConfirmationDialog(
  props
) {
  const {
    visible,
    deleteCollaborator,
    deletingCollaborator,
    data,
    onDismiss: onDismissProps,
  } = props;

  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="RemovePortalAdminConfirmationDialog.title" />
      ),
      subText: renderToString("RemovePortalAdminConfirmationDialog.message", {
        email: data?.email ?? "",
      }),
    };
  }, [data?.email, renderToString]);

  const onConfirmClicked = useCallback(() => {
    deleteCollaborator(data!.userID);
  }, [data, deleteCollaborator]);

  const onDismiss = useCallback(() => {
    if (!deletingCollaborator) {
      onDismissProps();
    }
  }, [onDismissProps, deletingCollaborator]);

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={dialogContentProps}
      modalProps={{ isBlocking: deletingCollaborator }}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <ButtonWithLoading
          onClick={onConfirmClicked}
          labelId="confirm"
          theme={themes.destructive}
          loading={deletingCollaborator}
          disabled={!visible}
        />
        <DefaultButton
          disabled={deletingCollaborator || !visible}
          onClick={onDismiss}
        >
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default RemovePortalAdminConfirmationDialog;
