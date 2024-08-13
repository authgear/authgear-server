import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useContext, useCallback, useMemo } from "react";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import { ConfirmationDialogStore } from "../../hook/useConfirmationDialog";

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

    const dialogContentProps = useMemo(() => {
      switch (action) {
        case MFAGracePeriodAction.Extend:
          return {
            title: (
              <FormattedMessage id="UserDetails.account-security.extend-mfa-grace-period-confirm-dialog.title" />
            ),
            subText: renderToString(
              "UserDetails.account-security.extend-mfa-grace-period-confirm-dialog.message"
            ),
          };
        case MFAGracePeriodAction.Grant:
        default:
          return {
            title: (
              <FormattedMessage id="UserDetails.account-security.grant-mfa-grace-period-confirm-dialog.title" />
            ),
            subText: renderToString(
              "UserDetails.account-security.grant-mfa-grace-period-confirm-dialog.message"
            ),
          };
      }
    }, [action, renderToString]);

    return (
      <Dialog
        hidden={!store.visible}
        dialogContentProps={dialogContentProps}
        modalProps={{ isBlocking: store.loading }}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <ButtonWithLoading
            onClick={onConfirmClicked}
            labelId="confirm"
            loading={store.loading ?? false}
          />
          <DefaultButton
            disabled={store.loading}
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };
