import React, { useContext, useMemo } from "react";
import {
  IDialogContentProps,
  Dialog,
  DialogType,
  DialogFooter,
  PrimaryButton,
  DefaultButton,
  IDialogProps,
  IButtonProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

export interface SwitchTabBlockerDialogProps extends IDialogProps {
  onDialogConfirm?: IButtonProps["onClick"];
  onDialogDismiss?: IButtonProps["onClick"];
}

const SwitchTabBlockerDialog: React.FC<SwitchTabBlockerDialogProps> = function SwitchTabBlockerDialog(
  props
) {
  const { onDialogConfirm, onDialogDismiss, ...rest } = props;

  const { renderToString } = useContext(Context);

  const dialogContentProps: IDialogContentProps = useMemo(
    () => ({
      type: DialogType.normal,
      title: <FormattedMessage id="SwitchTabBlockerDialog.title" />,
      subText: renderToString("SwitchTabBlockerDialog.content"),
    }),
    [renderToString]
  );

  return (
    <Dialog dialogContentProps={dialogContentProps} {...rest}>
      <DialogFooter>
        <PrimaryButton onClick={onDialogConfirm}>
          <FormattedMessage id="confirm" />
        </PrimaryButton>
        <DefaultButton onClick={onDialogDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default SwitchTabBlockerDialog;
