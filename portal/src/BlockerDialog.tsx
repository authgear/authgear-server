import React, { useContext, useMemo } from "react";
import {
  IDialogContentProps,
  Dialog,
  DialogType,
  DialogFooter,
  DefaultButton,
  IDialogProps,
  IButtonProps,
} from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";

export interface BlockerDialogProps extends IDialogProps {
  contentTitleId: string;
  contentSubTextId: string;
  contentConfirmId?: string;
  contentCancelId?: string;
  onDialogConfirm?: IButtonProps["onClick"];
  onDialogDismiss?: () => void;
}

const BlockerDialog: React.FC<BlockerDialogProps> = function BlockerDialog(
  props
) {
  const {
    contentTitleId,
    contentSubTextId,
    contentConfirmId,
    contentCancelId,
    onDialogConfirm,
    onDialogDismiss,
    ...rest
  } = props;

  const { themes } = useSystemConfig();
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
        <PrimaryButton onClick={onDialogConfirm} theme={themes.destructive}>
          <FormattedMessage id={contentConfirmId ?? "confirm"} />
        </PrimaryButton>
        <DefaultButton onClick={onDialogDismiss}>
          <FormattedMessage id={contentCancelId ?? "cancel"} />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default BlockerDialog;
