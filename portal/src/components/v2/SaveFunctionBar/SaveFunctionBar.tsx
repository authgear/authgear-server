import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "../../../intl";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { useSystemConfig } from "../../../context/SystemConfigContext";
import { PrimaryButton } from "../Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../Button/SecondaryButton/SecondaryButton";
import FluentPrimaryButton from "../../../PrimaryButton";
import DefaultButton from "../../../DefaultButton";
import styles from "./SaveFunctionBar.module.css";
import { useSaveFunctionBarAlignment } from "./useSaveFunctionBarAlignment";

export interface SaveFunctionBarProps {
  className?: string;
  anchorRef?: React.RefObject<HTMLElement | null>;
}

export function SaveFunctionBar({
  className,
  anchorRef,
}: SaveFunctionBarProps): React.ReactElement | null {
  const { canReset, canSave, isDirty, isUpdating, onReset, onSave } =
    useFormContainerBaseContext();
  const alignStyle = useSaveFunctionBarAlignment(anchorRef);
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const [isDiscardDialogVisible, setIsDiscardDialogVisible] = useState(false);
  const onOpenDiscardDialog = useCallback(() => {
    setIsDiscardDialogVisible(true);
  }, []);
  const onDismissDiscardDialog = useCallback(() => {
    setIsDiscardDialogVisible(false);
  }, []);
  const onConfirmDiscard = useCallback(() => {
    onReset();
    setTimeout(() => setIsDiscardDialogVisible(false), 0);
  }, [onReset]);

  const discardDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="FormContainer.reset-dialog.title" />,
      subText: renderToString("FormContainer.reset-dialog.message"),
    };
  }, [renderToString]);

  if (!isDirty) {
    return null;
  }

  const fixedStyle =
    anchorRef != null && alignStyle != null ? alignStyle : undefined;

  return (
    <>
      <div
        className={cn(styles.root, className)}
        style={fixedStyle}
        role="region"
        aria-live="polite"
      >
        <div className={styles.message}>
          <InfoCircledIcon className={styles.messageIcon} aria-hidden />
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="SaveFunctionBar.message" />
          </Text>
        </div>
        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            disabled={!canReset}
            text={<FormattedMessage id="SaveFunctionBar.discard" />}
            onClick={onOpenDiscardDialog}
          />
          <PrimaryButton
            size="2"
            disabled={!canSave}
            loading={isUpdating}
            text={<FormattedMessage id="save" />}
            onClick={onSave}
          />
        </div>
      </div>
      <Dialog
        hidden={!isDiscardDialogVisible}
        dialogContentProps={discardDialogContentProps}
        onDismiss={onDismissDiscardDialog}
      >
        <DialogFooter>
          <FluentPrimaryButton
            onClick={onConfirmDiscard}
            theme={themes.destructive}
            text={<FormattedMessage id="FormContainer.reset-dialog.confirm" />}
          />
          <DefaultButton
            onClick={onDismissDiscardDialog}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </>
  );
}
