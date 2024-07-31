import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useContext, useCallback, useMemo, useState } from "react";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";

interface ConfirmationDialogStore {
  visible: boolean;
  loading?: boolean;
  show: () => void;
  dismiss: () => void;
  confirm: () => void;
}

interface MarkPasswordAsExpiredConfirmationDialogProps {
  store: ConfirmationDialogStore;
  isExpired: boolean;
  onConfirm: () => void;
}

export function useConfirmationDialog(): ConfirmationDialogStore {
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);

  const show = useCallback(() => {
    setVisible(true);
  }, []);

  const dismiss = useCallback(() => {
    setVisible(false);
  }, []);

  const confirm = useCallback(() => {
    setLoading(true);
  }, []);

  return useMemo(() => {
    return {
      visible,
      loading,
      show,
      dismiss,
      confirm,
    };
  }, [visible, loading, show, dismiss, confirm]);
}

export const MarkPasswordAsExpiredConfirmationDialog: React.VFC<MarkPasswordAsExpiredConfirmationDialogProps> =
  function MarkPasswordAsExpiredConfirmationDialog(
    props: MarkPasswordAsExpiredConfirmationDialogProps
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
