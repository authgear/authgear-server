import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useContext, useCallback, useMemo } from "react";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import { ConfirmationDialogStore } from "../../hook/useConfirmationDialog";

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

    const dialogContentProps = useMemo(() => {
      return {
        title: (
          <FormattedMessage id="UserDetails.account-security.cancel-mfa-grace-period-confirm-dialog.title" />
        ),
        subText: renderToString(
          "UserDetails.account-security.cancel-mfa-grace-period-confirm-dialog.message"
        ),
      };
    }, [renderToString]);

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
