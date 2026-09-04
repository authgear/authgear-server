import { Context, FormattedMessage } from "../../intl";
import React, { useContext, useCallback, useMemo } from "react";
import { ConfirmationDialogStore } from "../../hook/useConfirmationDialog";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

interface SetPasswordExpiredConfirmationDialogProps {
  store: ConfirmationDialogStore;
  isExpired: boolean;
  onConfirm: () => void;
}

export const SetPasswordExpiredConfirmationDialog: React.VFC<SetPasswordExpiredConfirmationDialogProps> =
  function SetPasswordExpiredConfirmationDialog(
    props: SetPasswordExpiredConfirmationDialogProps
  ) {
    const { store, isExpired, onConfirm: onConfirmProp } = props;

    const { renderToString } = useContext(Context);

    const onConfirmClicked = useCallback(() => {
      onConfirmProp();
    }, [onConfirmProp]);

    const onDismiss = useCallback(() => {
      if (!store.loading) {
        store.dismiss();
      }
    }, [store]);

    const description = useMemo(() => {
      return isExpired
        ? renderToString(
            "UserDetails.account-security.mark-as-expired-confirm-dialog.message.revoke"
          )
        : renderToString(
            "UserDetails.account-security.mark-as-expired-confirm-dialog.message"
          );
    }, [isExpired, renderToString]);

    return (
      <ConfirmationDialog
        open={store.visible}
        onOpenChange={(open) => {
          if (!open) {
            onDismiss();
          }
        }}
        title={
          <FormattedMessage id="UserDetails.account-security.mark-as-expired-confirm-dialog.title" />
        }
        description={description}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onDismiss}
        loading={store.loading}
        confirmColor="red"
      />
    );
  };
