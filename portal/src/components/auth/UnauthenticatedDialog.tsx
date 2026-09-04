import { Dialog } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import React from "react";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import styles from "./UnauthenticatedDialog.module.css";

interface UnauthenticatedDialogProps {
  isHidden: boolean;
  onConfirm: () => void;
}

export const UnauthenticatedDialog: React.VFC<UnauthenticatedDialogProps> =
  function UnauthenticatedDialog({ isHidden, onConfirm }) {
    // This is a blocking dialog: the user must re-authenticate, so it has no
    // close button and cannot be dismissed by Escape or clicking the overlay.
    // We keep it fully controlled via `isHidden` (no onOpenChange), and prevent
    // the built-in dismissal gestures below.
    const preventDismiss = React.useCallback((e: Event) => {
      e.preventDefault();
    }, []);

    return (
      <Dialog.Root open={!isHidden}>
        <Dialog.Content
          maxWidth="400px"
          size="3"
          onEscapeKeyDown={preventDismiss}
          onPointerDownOutside={preventDismiss}
          onInteractOutside={preventDismiss}
        >
          <Dialog.Title>
            <FormattedMessage id="UnauthenticatedDialog.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            <FormattedMessage id="UnauthenticatedDialog.description" />
          </Dialog.Description>
          <div className={styles.actions}>
            <PrimaryButton
              size="2"
              onClick={onConfirm}
              text={<FormattedMessage id="UnauthenticatedDialog.button" />}
            />
          </div>
        </Dialog.Content>
      </Dialog.Root>
    );
  };
