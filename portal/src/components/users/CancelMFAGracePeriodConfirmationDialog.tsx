import { Context, FormattedMessage } from "../../intl";
import React, { useContext, useCallback, useMemo } from "react";
import { ConfirmationDialogStore } from "../../hook/useConfirmationDialog";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

interface CancelMFAGracePeriodConfirmationDialogProps {
  store: ConfirmationDialogStore;
  onConfirm: () => void;
}

export const CancelMFAGracePeriodConfirmationDialog: React.VFC<CancelMFAGracePeriodConfirmationDialogProps> =
  function CancelMFAGracePeriodConfirmationDialog(
    props: CancelMFAGracePeriodConfirmationDialogProps
  ) {
    const { store, onConfirm: onConfirmProp } = props;

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
      return renderToString(
        "UserDetails.account-security.cancel-mfa-grace-period-confirm-dialog.message"
      );
    }, [renderToString]);

    return (
      <ConfirmationDialog
        open={store.visible}
        onOpenChange={(open) => {
          if (!open) {
            onDismiss();
          }
        }}
        title={
          <FormattedMessage id="UserDetails.account-security.cancel-mfa-grace-period-confirm-dialog.title" />
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
