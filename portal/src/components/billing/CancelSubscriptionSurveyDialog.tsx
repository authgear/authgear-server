import {
  Dialog,
  DialogFooter,
  DialogType,
  IDialogContentProps,
} from "@fluentui/react";
import React, { useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

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
  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      type: DialogType.normal,
      title: <FormattedMessage id="CancelSubscriptionSurveyDialog.title" />,
      subText: (
        <FormattedMessage id="CancelSubscriptionSurveyDialog.body" />
      ) as unknown as IDialogContentProps["subText"],
    };
  }, []);

  return (
    <Dialog
      hidden={isHidden}
      onDismiss={onDismiss}
      dialogContentProps={dialogContentProps}
    >
      <DialogFooter>
        <PrimaryButton
          onClick={onConfirm}
          disabled={isHidden}
          text={
            <FormattedMessage id="CancelSubscriptionSurveyDialog.button.yes" />
          }
        />
        <DefaultButton
          onClick={onCancel}
          text={
            <FormattedMessage id="CancelSubscriptionSurveyDialog.button.no" />
          }
        />
      </DialogFooter>
    </Dialog>
  );
}
