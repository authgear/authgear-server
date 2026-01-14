import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import React, { useContext, useCallback, useMemo } from "react";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import { ConfirmationDialogStore } from "../../hook/useConfirmationDialog";

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

    const removeConfirmDialogContentProps = useMemo(() => {
      return {
        title: (
          <FormattedMessage id="UserDetails.account-security.mark-as-expired-confirm-dialog.title" />
        ),
        subText: isExpired
          ? renderToString(
              "UserDetails.account-security.mark-as-expired-confirm-dialog.message.revoke"
            )
          : renderToString(
              "UserDetails.account-security.mark-as-expired-confirm-dialog.message"
            ),
      };
    }, [isExpired, renderToString]);

    return (
      <Dialog
        hidden={!store.visible}
        dialogContentProps={removeConfirmDialogContentProps}
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
