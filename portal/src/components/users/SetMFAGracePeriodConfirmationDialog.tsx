import { Context, FormattedMessage } from "../../intl";
import React, { useContext, useCallback, useMemo } from "react";
import { ConfirmationDialogStore } from "../../hook/useConfirmationDialog";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

export enum MFAGracePeriodAction {
  Grant = "grant",
  Extend = "extend",
}

interface SetMFAGracePeriodConfirmationDialogProps {
  store: ConfirmationDialogStore;
  action: MFAGracePeriodAction;
  onConfirm: () => void;
}

export const SetMFAGracePeriodConfirmationDialog: React.VFC<SetMFAGracePeriodConfirmationDialogProps> =
  function SetMFAGracePeriodConfirmationDialog(
    props: SetMFAGracePeriodConfirmationDialogProps
  ) {
    const { store, action, onConfirm: onConfirmProp } = props;

    const { renderToString } = useContext(Context);

    const onConfirmClicked = useCallback(() => {
      onConfirmProp();
    }, [onConfirmProp]);

    const onDismiss = useCallback(() => {
      if (!store.loading) {
        store.dismiss();
      }
    }, [store]);

    const dialogContent = useMemo(() => {
      switch (action) {
        case MFAGracePeriodAction.Extend:
          return {
            title: (
              <FormattedMessage id="UserDetails.account-security.extend-mfa-grace-period-confirm-dialog.title" />
            ),
            description: renderToString(
              "UserDetails.account-security.extend-mfa-grace-period-confirm-dialog.message"
            ),
          };
        case MFAGracePeriodAction.Grant:
        default:
          return {
            title: (
              <FormattedMessage id="UserDetails.account-security.grant-mfa-grace-period-confirm-dialog.title" />
            ),
            description: renderToString(
              "UserDetails.account-security.grant-mfa-grace-period-confirm-dialog.message"
            ),
          };
      }
    }, [action, renderToString]);

    return (
      <ConfirmationDialog
        open={store.visible}
        onOpenChange={(open) => {
          if (!open) {
            onDismiss();
          }
        }}
        title={dialogContent.title}
        description={dialogContent.description}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onDismiss}
        loading={store.loading}
        confirmColor="red"
      />
    );
  };
