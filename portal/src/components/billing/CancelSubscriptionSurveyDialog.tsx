import { Dialog } from "@radix-ui/themes";
import React from "react";
import { FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import styles from "./CancelSubscriptionSurveyDialog.module.css";

interface CancelSubscriptionSurveyDialogProps {
  isHidden: boolean;
  onDismiss: () => void;
  onConfirm: () => void;
  onCancel: () => void;
}

export function CancelSubscriptionSurveyDialog({
  isHidden,
  onDismiss,
  onConfirm,
  onCancel,
}: CancelSubscriptionSurveyDialogProps): React.ReactElement {
  return (
    <Dialog.Root
      open={!isHidden}
      onOpenChange={(open) => {
        if (!open) {
          onDismiss();
        }
      }}
    >
      <Dialog.Content maxWidth="400px" size="3">
        <Dialog.Title>
          <FormattedMessage id="CancelSubscriptionSurveyDialog.title" />
        </Dialog.Title>
        <Dialog.Description size="2">
          <FormattedMessage id="CancelSubscriptionSurveyDialog.body" />
        </Dialog.Description>
        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            onClick={onCancel}
            text={
              <FormattedMessage id="CancelSubscriptionSurveyDialog.button.no" />
            }
          />
          <PrimaryButton
            size="2"
            onClick={onConfirm}
            disabled={isHidden}
            text={
              <FormattedMessage id="CancelSubscriptionSurveyDialog.button.yes" />
            }
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
