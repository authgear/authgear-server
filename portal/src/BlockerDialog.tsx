import React, { useContext, useMemo } from "react";
import {
  IDialogContentProps,
  Dialog,
  DialogType,
  DialogFooter,
  IDialogProps,
  IButtonProps,
} from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import { Context, FormattedMessage } from "./intl";
import { useSystemConfig } from "./context/SystemConfigContext";

export interface BlockerDialogProps extends IDialogProps {
  contentTitleId: string;
  contentSubTextId: string;
  contentConfirmId?: string;
  contentCancelId?: string;
  onDialogConfirm?: IButtonProps["onClick"];
  onDialogDismiss?: () => void;
}

const BlockerDialog: React.VFC<BlockerDialogProps> = function BlockerDialog(
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
        <PrimaryButton
          onClick={onDialogConfirm}
          theme={themes.destructive}
          text={<FormattedMessage id={contentConfirmId ?? "confirm"} />}
        />
        <DefaultButton
          onClick={onDialogDismiss}
          text={<FormattedMessage id={contentCancelId ?? "cancel"} />}
        />
      </DialogFooter>
    </Dialog>
  );
};

export default BlockerDialog;
