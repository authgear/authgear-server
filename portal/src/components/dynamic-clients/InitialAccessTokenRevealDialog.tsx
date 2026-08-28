import React, { useCallback } from "react";
import { Dialog } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { Callout } from "../v2/Callout/Callout";
import { TextField } from "../v2/TextField/TextField";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import styles from "./InitialAccessTokenRevealDialog.module.css";

export interface InitialAccessTokenRevealDialogProps {
  // The plaintext token to reveal; the dialog is open while this is non-null.
  token: string | null;
  onDismiss: () => void;
}

export function InitialAccessTokenRevealDialog({
  token,
  onDismiss,
}: InitialAccessTokenRevealDialogProps): React.ReactElement {
  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        onDismiss();
      }
    },
    [onDismiss]
  );

  return (
    <Dialog.Root open={token != null} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage id="InitialAccessTokenRevealDialog.title" />
        </Dialog.Title>
        <Dialog.Description size="2" color="gray">
          <FormattedMessage id="InitialAccessTokenRevealDialog.description" />
        </Dialog.Description>
        <div className={styles.content}>
          <TextField
            size="2"
            label={
              <FormattedMessage id="InitialAccessTokenRevealDialog.token.label" />
            }
            value={token ?? ""}
            readOnly={true}
            suffixPlain={true}
            suffix={<CopyIconButton textToCopy={token ?? ""} />}
          />
          <Callout
            type="warning"
            text={
              <FormattedMessage id="InitialAccessTokenRevealDialog.warning" />
            }
          />
        </div>
        <div className={styles.actions}>
          <PrimaryButton
            size="2"
            text={<FormattedMessage id="InitialAccessTokenRevealDialog.done" />}
            onClick={onDismiss}
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
