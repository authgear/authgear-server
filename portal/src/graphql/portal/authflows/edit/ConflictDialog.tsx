import React, { useState } from "react";
import { Button, Dialog } from "@radix-ui/themes";
import { FormattedMessage } from "../../../../intl";
import { useFormConflictErrors } from "../../../../form";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import styles from "./ConflictDialog.module.css";

export interface ConflictDialogProps {
  // Load the latest config and keep the flow being edited on top of it.
  onReload: () => void;
  // Save again over the other changes.
  onOverwrite: () => void;
}

// Shown when a save finds the project changed since it was loaded.
export function ConflictDialog({
  onReload,
  onOverwrite,
}: ConflictDialogProps): React.ReactElement {
  const isConflicted = useFormConflictErrors().length !== 0;
  // Open when a save conflicts; closed by any action, or once resolved.
  const [visible, setVisible] = useState(isConflicted);
  const [prevIsConflicted, setPrevIsConflicted] = useState(isConflicted);
  if (prevIsConflicted !== isConflicted) {
    setPrevIsConflicted(isConflicted);
    setVisible(isConflicted);
  }
  return (
    <Dialog.Root open={visible} onOpenChange={setVisible}>
      <Dialog.Content maxWidth="440px" size="3">
        <Dialog.Title>
          <FormattedMessage id="AuthFlowsScreen.conflict.title" />
        </Dialog.Title>
        <Dialog.Description size="2">
          <FormattedMessage id="AuthFlowsScreen.conflict.description" />
        </Dialog.Description>
        <div className={styles.actions}>
          <Button
            size="2"
            variant="soft"
            color="red"
            onClick={() => {
              setVisible(false);
              onOverwrite();
            }}
          >
            <FormattedMessage id="AuthFlowsScreen.conflict.overwrite" />
          </Button>
          <span className={styles.spacer} />
          <SecondaryButton
            size="2"
            text={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
            onClick={() => setVisible(false)}
          />
          <Button
            size="2"
            onClick={() => {
              setVisible(false);
              onReload();
            }}
          >
            <FormattedMessage id="AuthFlowsScreen.conflict.reload" />
          </Button>
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
