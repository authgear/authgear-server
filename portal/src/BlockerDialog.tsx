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

export interface BlockerDialogProps extends IDialogProps {
  contentTitleId: string;
  contentSubTextId: string;
  onDialogConfirm?: IButtonProps["onClick"];
  onDialogDismiss?: () => void;
}

const BlockerDialog: React.FC<BlockerDialogProps> = function BlockerDialog(
  props
) {
  const {
    contentTitleId,
    contentSubTextId,
    onDialogConfirm,
    onDialogDismiss,
    ...rest
  } = props;

  const { renderToString } = useContext(Context);

  const dialogContentProps: IDialogContentProps = useMemo(
    () => ({
      type: DialogType.normal,
      title: <FormattedMessage id={contentTitleId} />,
      subText: renderToString(contentSubTextId),
    }),
    [renderToString, contentTitleId, contentSubTextId]
  );

  return (
    <Dialog
      dialogContentProps={dialogContentProps}
      onDismiss={onDialogDismiss}
      {...rest}
    >
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

export default BlockerDialog;
